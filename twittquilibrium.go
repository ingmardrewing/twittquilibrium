package twittquilibrium

import (
	"fmt"
	"log"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

func NewTwittquilibrium(accessToken, accessTokenSecret, consumerKey, consumerKeySecret string) *twittquilibrium {
	tq := new(twittquilibrium)
	tq.accessToken = accessToken
	tq.accessTokenSecret = accessTokenSecret
	tq.consumerKey = consumerKey
	tq.consumerKeySecret = consumerKeySecret
	tq.exceptUsers = map[string]bool{}
	tq.disposableUsers = []twitter.User{}
	tq.init()
	return tq
}

type twittquilibrium struct {
	disposableUsers   []twitter.User
	accessToken       string
	accessTokenSecret string
	consumerKey       string
	consumerKeySecret string
	exceptUsers       map[string]bool
	client            *twitter.Client
}

func (t *twittquilibrium) init() {
	config := oauth1.NewConfig(t.consumerKey, t.consumerKeySecret)
	token := oauth1.NewToken(t.accessToken, t.accessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	t.client = twitter.NewClient(httpClient)
}

func (t *twittquilibrium) KeepFollowing(userHandle string) {
	t.exceptUsers[userHandle] = true
}

func (t *twittquilibrium) Clean() {
	t.RetrieveFollowedUsers()
	t.AddFollwersToBeKept()
	t.AddVerifiedUsersToBeKept()
	t.DisposeOfTheRest()
}

// Retreive followed users
func (t *twittquilibrium) RetrieveFollowedUsers() {
	var cursor int64
LP:
	for {
		friends, _, err := t.client.Friends.List(&twitter.FriendListParams{Count: 200, Cursor: cursor})
		if err != nil {
			log.Fatalln(err)
		}
		if len(friends.Users) < 200 {
			break LP
		}
		t.disposableUsers = append(t.disposableUsers, friends.Users...)
		cursor = friends.NextCursor
	}
}

// Keep following users, who are following back
func (t *twittquilibrium) AddFollwersToBeKept() {
	var cursor int64 = -1

	usrs := []twitter.User{}
LP:
	for {
		followers, _, err := t.client.Followers.List(&twitter.FollowerListParams{Count: 200, Cursor: cursor})
		if err != nil {
			log.Fatalln(err)
		}
		usrs = append(usrs, followers.Users...)
		if len(followers.Users) < 200 {
			break LP
		}
		cursor = followers.NextCursor
	}
	for _, u := range usrs {
		t.KeepFollowing(u.ScreenName)
	}
}

// Keep following official accounts
func (t *twittquilibrium) AddVerifiedUsersToBeKept() {
	for _, u := range t.disposableUsers {
		if u.Verified {
			t.KeepFollowing(u.ScreenName)
		}
	}
}

// Dispose of unwanted followed users
func (t *twittquilibrium) DisposeOfTheRest() {
	for _, disposableUser := range t.disposableUsers {
		if ok, _ := t.exceptUsers[disposableUser.ScreenName]; !ok {
			removedUser, _, err := t.client.Friendships.Destroy(&twitter.FriendshipDestroyParams{disposableUser.ScreenName, disposableUser.ID})
			if err != nil {
				log.Fatalln(err)
			}
			now := time.Now()
			timeStamp := now.Format("[2006-01-02 15:04]")
			fmt.Printf(timeStamp+" unfollowed user %s\n",
				removedUser.ScreenName)
		}
	}
}
