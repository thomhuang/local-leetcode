package src

func ToQuestionMap(response AllQuestionsResponse) map[int]string {
	if len(response.Response) == 0 {
		return map[int]string{}
	}

	res := make(map[int]string)
	for _, question := range response.Response {
		res[question.Metadata.QuestionId] = question.Metadata.QuestionTitleSlug
	}

	return res
}

func ToQuestion(response QuestionResponse) Question {
	return Question{}
}
