//go:build inefficient_feed

package design_twitter

import "container/heap"

/*
	Followings:
		Represented by a map of maps, easy to add/delete users
	Posts:
		Will keep a map -> posts entity, in order
	Feed:
		Go through users following map, and fetch posts from the users
		that they follow (including themselves) and push to a heap
		that is sorted by the post timestamp
*/

type Post struct {
	postId, tweetId, userId int
}

type FeedHeap []Post

func (h FeedHeap) Len() int {
	return len(h)
}

func (h FeedHeap) Less(i, j int) bool {
	return h[i].postId > h[j].postId
}

func (h FeedHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *FeedHeap) Push(x any) {
	*h = append(*h, x.(Post))
}

func (h *FeedHeap) Pop() any {
	n := len(*h)

	popped := (*h)[n-1]
	*h = (*h)[:n-1]

	return popped
}

type Twitter struct {
	Following  map[int]map[int]bool
	Posts      map[int][]Post
	lastPostId int
}

func Constructor() Twitter {
	return Twitter{
		Following:  make(map[int]map[int]bool),
		Posts:      make(map[int][]Post),
		lastPostId: 0,
	}

}

// Create tweet for userId with tweetId; all tweetId is unique
func (this *Twitter) PostTweet(userId int, tweetId int) {
	// Add the post to the user's map of posts ...
	this.Posts[userId] = append(this.Posts[userId], Post{postId: this.lastPostId, tweetId: tweetId, userId: userId})
	// increment our postId that we use to determine "latest" posts
	this.lastPostId++
}

// 10 most recent tweet IDs in user's news feed
// either their posts, or posts of users they follow
func (this *Twitter) GetNewsFeed(userId int) []int {
	feed := &FeedHeap{}
	// add all the user's posts first
	for _, post := range this.Posts[userId] {
		heap.Push(feed, post)
	}

	// add all of the users followings
	followings := this.Following[userId]
	for following := range followings {
		if following == userId {
			continue
		}

		userPosts := this.Posts[following]

		for _, post := range userPosts {
			heap.Push(feed, post)
		}
	}

	feedLength := min(feed.Len(), 10)
	res := make([]int, feedLength)
	for i := range feedLength {
		res[i] = heap.Pop(feed).(Post).tweetId
	}

	return res
}

// follow a user
func (this *Twitter) Follow(followerId int, followeeId int) {
	if this.Following == nil {
		this.Following = make(map[int]map[int]bool)
	}

	followings, exists := this.Following[followerId]

	if !exists {
		followings = make(map[int]bool)
		this.Following[followerId] = followings
	}

	followings[followeeId] = true
}

// unfollow a user
func (this *Twitter) Unfollow(followerId int, followeeId int) {
	followings, exists := this.Following[followerId]

	if !exists {
		return
	}

	delete(followings, followeeId)
}

/**
 * Your Twitter object will be instantiated and called as such:
 * obj := Constructor();
 * obj.PostTweet(userId,tweetId);
 * param_2 := obj.GetNewsFeed(userId);
 * obj.Follow(followerId,followeeId);
 * obj.Unfollow(followerId,followeeId);
 */
