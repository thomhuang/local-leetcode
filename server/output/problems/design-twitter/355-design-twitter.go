package design_twitter

type Twitter struct {
}

func Constructor() Twitter {
	panic("not implemented")

}

func (this *Twitter) PostTweet(userId int, tweetId int) {

}

func (this *Twitter) GetNewsFeed(userId int) []int {
	panic("not implemented")

}

func (this *Twitter) Follow(followerId int, followeeId int) {

}

func (this *Twitter) Unfollow(followerId int, followeeId int) {

}

/**
 * Your Twitter object will be instantiated and called as such:
 * obj := Constructor();
 * obj.PostTweet(userId,tweetId);
 * param_2 := obj.GetNewsFeed(userId);
 * obj.Follow(followerId,followeeId);
 * obj.Unfollow(followerId,followeeId);
 */
