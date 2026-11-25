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

type FeedItem struct {
	post      Post
	postIndex int
}

type FeedHeap []FeedItem

func (h FeedHeap) Len() int {
	return len(h)
}

func (h FeedHeap) Less(i, j int) bool {
	return h[i].post.postId > h[j].post.postId
}

func (h FeedHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *FeedHeap) Push(x any) {
	*h = append(*h, x.(FeedItem))
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
	// merge K sorted lists approach ...
	followings, exists := this.Following[userId]

	// if user hasn't followed/unfollowed anyone so far,
	// initialize their followings of themselves ...
	if !exists {
		followings = this.InitializeUser(userId)
	}

	feed := &FeedHeap{}

	for id := range followings {
		posts := this.Posts[id]
		if len(posts) == 0 {
			continue
		}

		// push the users last created post to our heap
		lastIndex := len(posts) - 1
		heap.Push(feed, FeedItem{
			post:      posts[lastIndex],
			postIndex: lastIndex,
		})
	}

	res := []int{}

	// while we have user posts to process, and we have
	// posts to add to our feed of length 10
	for feed.Len() > 0 && len(res) < 10 {
		item := heap.Pop(feed).(FeedItem)
		// the last added post is guaranteed to be at the front
		// as we went through all followed users and added their last added post
		res = append(res, item.post.tweetId)

		// if not the first item, there are posts to processs
		if item.postIndex != 0 {
			// push the user's next post
			previousIndex := item.postIndex - 1
			heap.Push(feed, FeedItem{
				post:      this.Posts[item.post.userId][previousIndex],
				postIndex: previousIndex,
			})
		}
	}

	return res
}

// follow a user
func (this *Twitter) Follow(followerId int, followeeId int) {
	followings, exists := this.Following[followerId]

	// if user hasn't followed anyone yet, initialize their map
	if !exists {
		followings = this.InitializeUser(followerId)
	}

	followings[followeeId] = true
}

// unfollow a user
func (this *Twitter) Unfollow(followerId int, followeeId int) {
	followings, exists := this.Following[followerId]

	// if user hasn't followed anyone yet, no one to unfollow ... return
	if !exists {
		followings = this.InitializeUser(followerId)
		return
	}

	// delete item out of map
	delete(followings, followeeId)
}

func (this *Twitter) InitializeUser(followerId int) map[int]bool {
	followings := make(map[int]bool)
	followings[followerId] = true
	this.Following[followerId] = followings

	return followings
}

/**
 * Your Twitter object will be instantiated and called as such:
 * obj := Constructor();
 * obj.PostTweet(userId,tweetId);
 * param_2 := obj.GetNewsFeed(userId);
 * obj.Follow(followerId,followeeId);
 * obj.Unfollow(followerId,followeeId);
 */
