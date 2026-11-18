package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/thomhuang/local-leetcode/internal/question"
)

// loadQuestions builds the frontend-ID-to-slug map from the cached problem list,
// local problem sets, and downloaded problem directories. Unlike the old 5-day
// REST refresh, problem IDs and slugs are immutable so an existing cache file
// never goes stale, and unknown IDs are resolved on demand via GraphQL.
func (app *App) loadQuestions() map[int]question.QuestionMetadata {
	return app.loadQuestionsFrom(allProblemsFile, problemSetsDir, problemsDir)
}

func (app *App) loadQuestionsFrom(cachePath, setsDir, probsDir string) map[int]question.QuestionMetadata {
	res := make(map[int]question.QuestionMetadata)

	if data, err := os.ReadFile(cachePath); err == nil {
		var cached question.AllQuestionsResponse
		if err := json.Unmarshal(data, &cached); err != nil {
			app.log().Warn("questions.cache", "unable to parse cached problem list", "path", cachePath, "error", err.Error())
		} else {
			res = question.ToQuestionMap(cached)
			app.log().Debug("questions.cache", "loaded cached problem list", "path", cachePath, "problems", len(res))
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		app.log().Warn("questions.cache", "unable to read cached problem list", "path", cachePath, "error", err.Error())
	}

	mergeProblemSetsInto(res, setsDir)
	mergeLocalProblemsInto(res, probsDir)

	app.log().Info("questions.loaded", "problem list ready", "problems", len(res))
	return res
}

func mergeProblemSetsInto(dest map[int]question.QuestionMetadata, setsDir string) {
	entries, err := os.ReadDir(setsDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(setsDir + "/" + entry.Name())
		if err != nil {
			continue
		}
		set, err := question.ParseProblemSet(data)
		if err != nil {
			continue
		}
		for _, p := range set.Problems() {
			if p.ID <= 0 || p.Slug == "" {
				continue
			}
			if existing, ok := dest[p.ID]; !ok || existing.QuestionTitleSlug == "" {
				dest[p.ID] = question.QuestionMetadata{
					FrontEndQuestionId: p.ID,
					QuestionTitleSlug:  p.Slug,
				}
			}
		}
	}
}

func mergeLocalProblemsInto(dest map[int]question.QuestionMetadata, probsDir string) {
	dirs, err := os.ReadDir(probsDir)
	if err != nil {
		return
	}
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		slug := dir.Name()
		files, err := os.ReadDir(probsDir + "/" + slug)
		if err != nil {
			continue
		}
		suffix := "-" + slug + ".go"
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), suffix) {
				continue
			}
			idStr := strings.TrimSuffix(f.Name(), suffix)
			id, err := strconv.Atoi(idStr)
			if err != nil || id <= 0 {
				continue
			}
			if existing, ok := dest[id]; !ok || existing.QuestionTitleSlug == "" {
				dest[id] = question.QuestionMetadata{
					FrontEndQuestionId: id,
					QuestionTitleSlug:  slug,
				}
			}
		}
	}
}

// resolveQuestion looks up a problem by its frontend number, falling back to a
// live GraphQL search when the problem is not yet in the local map.
func (app *App) resolveQuestion(id int) (question.QuestionMetadata, error) {
	if app != nil && app.Questions != nil {
		if meta, ok := app.Questions[id]; ok && meta.QuestionTitleSlug != "" {
			return meta, nil
		}
	}
	if app == nil || app.client == nil {
		return question.QuestionMetadata{}, fmt.Errorf("problem %d was not found in the problem list", id)
	}

	app.log().Info("questions.lookup", "searching GraphQL for problem", "problem", id)
	resp, err := app.fetchProblemsetSearch(strconv.Itoa(id))
	if err != nil {
		app.log().Error("questions.lookup", "search failed", "problem", id, "error", err.Error())
		return question.QuestionMetadata{}, fmt.Errorf("lookup problem %d: %w", id, err)
	}

	wantID := strconv.Itoa(id)
	for _, node := range resp.Data.ProblemsetQuestionList.Questions {
		if strings.TrimSpace(node.FrontEndQuestionId) == wantID && node.TitleSlug != "" {
			meta := question.QuestionMetadata{
				FrontEndQuestionId: id,
				QuestionTitleSlug:  node.TitleSlug,
			}
			if app.Questions == nil {
				app.Questions = make(map[int]question.QuestionMetadata)
			}
			app.Questions[id] = meta
			if app.baseURL == "" {
				app.saveQuestionsCache(allProblemsFile)
			}
			app.log().Info("questions.lookup", "resolved problem", "problem", id, "slug", node.TitleSlug)
			return meta, nil
		}
	}

	return question.QuestionMetadata{}, fmt.Errorf("problem %d was not found in the problem list", id)
}

func (app *App) saveQuestionsCache(cachePath string) {
	if len(app.Questions) == 0 {
		return
	}
	ids := make([]int, 0, len(app.Questions))
	for id := range app.Questions {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	pairs := make([]question.QuestionStatPair, 0, len(ids))
	for _, id := range ids {
		pairs = append(pairs, question.QuestionStatPair{Metadata: app.Questions[id]})
	}
	data, err := json.Marshal(question.AllQuestionsResponse{Response: pairs})
	if err != nil {
		app.log().Warn("questions.cache", "encode cache failed", "error", err.Error())
		return
	}
	if err := os.WriteFile(cachePath, data, filePerm); err != nil {
		app.log().Warn("questions.cache", "write cache failed", "path", cachePath, "error", err.Error())
	}
}
