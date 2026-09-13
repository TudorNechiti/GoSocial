package db

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/nechititudorr/GoSocial/internal/store"
)

var usernames = []string{
	"Tuds",
	"Alex",
	"Illia",
	"Nickiy",
	"Mara",
	"Bogdan",
	"Elena",
	"Radu",
	"Ioana",
	"Cristi",
	"Diana",
	"Vlad",
	"Simona",
	"Andrei",
	"Larisa",
	"Mihai",
	"Ana",
	"Stefan",
	"Roxana",
	"Dragos",
	"Alina",
	"Cosmin",
	"Ovidiu",
	"Georgiana",
	"Florin",
}

var titles = []string{
	"Getting Started with Go",
	"Understanding Context in Go",
	"Building REST APIs the Simple Way",
	"Why Postgres Is Still Great",
	"Docker for Beginners",
	"A Guide to Database Migrations",
	"Writing Clean Middleware",
	"Debugging Like a Pro",
	"The Power of Interfaces",
	"Concurrency Made Easy",
	"Structuring Go Projects",
	"Working with JSON in Go",
	"Error Handling Best Practices",
	"Testing Your Go Code",
	"Deploying Your First App",
	"Understanding Goroutines",
	"SQL Joins Explained",
	"Building a Blog API",
	"Authentication 101",
	"Tips for Faster Queries",
}

var contents = []string{
	"This is a sample post used for seeding the database.",
	"Just testing out some content here, nothing fancy.",
	"Go makes it really easy to build fast APIs.",
	"Not sure what to write, but this fills the gap.",
	"Here's some placeholder text for demo purposes.",
	"Learning something new about databases today.",
	"Docker and Postgres make local dev so much easier.",
	"This post exists purely for seed data.",
	"A quick thought on writing clean code.",
	"Random content generated for testing pagination.",
}

func Seed(store store.Storage) error {
	ctx := context.Background()

	users := generateUsers(100)
	for _, user := range users {
		if err := store.Users.Create(ctx, user); err != nil {
			log.Println("Error creating user:", err)
			return err
		}
	}

	posts := generatePosts(200, users)
	for _, post := range posts {
		if err := store.Posts.Create(ctx, post); err != nil {
			log.Println("Error creating post:", err)
			return err
		}
	}

	comments := generateComments(500, users, posts)
	for _, comment := range comments {
		if err := store.Comments.Create(ctx, comment); err != nil {
			log.Println("Error creating comment:", err)
			return err
		}
	}

	log.Println("Seeding complete!")
	return nil
}

func generateUsers(num int) []*store.User {
	users := make([]*store.User, num)

	for i := 0; i < num; i++ {
		users[i] = &store.User{
			Username: usernames[rand.Intn(len(usernames))] + fmt.Sprintf("%d", i),
			Email:    usernames[i%len(usernames)] + fmt.Sprintf("%d", i) + "@example.com",
			Password: "123123",
		}
	}

	return users
}

func generatePosts(num int, users []*store.User) []*store.Post {
	posts := make([]*store.Post, num)
	for i := 0; i < num; i++ {
		user := users[rand.Intn(len(users))]

		posts[i] = &store.Post{
			UserID:  user.ID,
			Title:   titles[rand.Intn(len(titles))],
			Content: contents[rand.Intn(len(contents))],
			Tags: []string{
				titles[rand.Intn(len(titles))],
				titles[rand.Intn(len(titles))],
			},
		}
	}

	return posts
}

func generateComments(num int, users []*store.User, posts []*store.Post) []*store.Comment {
	comments := make([]*store.Comment, num)
	for i := 0; i < num; i++ {
		comments[i] = &store.Comment{
			UserID:  users[rand.Intn(len(users))].ID,
			PostID:  posts[rand.Intn(len(posts))].ID,
			Content: contents[rand.Intn(len(contents))],
		}
	}

	return comments
}
