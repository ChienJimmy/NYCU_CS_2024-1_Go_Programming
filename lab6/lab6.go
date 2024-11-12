package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type Book struct {
	// TODO: Finish struct
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Pages int    `json:"pages"`
}

var bookshelf = []Book{
	// TODO: Init bookshelf
	{ID: 1, Name: "Blue Bird", Pages: 500},
}

// Check if a book name already exists
func isDuplicateName(name string) bool {
	for _, book := range bookshelf {
		if book.Name == name {
			return true
		}
	}
	return false
}

// Track the last used ID
var lastID = 1

func getBooks(c *gin.Context) {
	c.JSON(http.StatusOK, bookshelf)
}
func getBook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid ID"})
		return
	}

	for _, book := range bookshelf {
		if book.ID == id {
			c.JSON(http.StatusOK, book)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"message": "book not found"})
}
func addBook(c *gin.Context) {
	var newBook Book
	if err := c.ShouldBindJSON(&newBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid input"})
		return
	}

	// Check for duplicate book name (case-sensitive)
	if isDuplicateName(newBook.Name) {
		c.JSON(http.StatusConflict, gin.H{"message": "duplicate book name"})
		return
	}

	// Generate a new ID by incrementing `lastID`
	lastID++
	newBook.ID = lastID
	bookshelf = append(bookshelf, newBook)
	c.JSON(http.StatusCreated, newBook)
}
func deleteBook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid ID"})
		return
	}

	for i, book := range bookshelf {
		if book.ID == id {
			bookshelf = append(bookshelf[:i], bookshelf[i+1:]...)
			c.Status(http.StatusNoContent) // Return 204 No Content
			return
		}
	}
	c.Status(http.StatusNoContent) // Return 204 No Content even if the book is not found
}
func updateBook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid ID"})
		return
	}

	var updatedBook Book
	if err := c.ShouldBindJSON(&updatedBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid input"})
		return
	}

	// Check for duplicate book name (case-sensitive)
	for _, book := range bookshelf {
		if book.Name == updatedBook.Name && book.ID != id {
			c.JSON(http.StatusConflict, gin.H{"message": "duplicate book name"})
			return
		}
	}

	for i, book := range bookshelf {
		if book.ID == id {
			updatedBook.ID = id // Ensure the ID remains unchanged
			bookshelf[i] = updatedBook
			c.JSON(http.StatusOK, updatedBook)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"message": "book not found"})
}

func main() {
	r := gin.Default()
	r.RedirectFixedPath = true

	// TODO: Add routes
	r.GET("/bookshelf", getBooks)
	r.GET("/bookshelf/:id", getBook)
	r.POST("/bookshelf", addBook)
	r.DELETE("/bookshelf/:id", deleteBook)
	r.PUT("/bookshelf/:id", updateBook)

	err := r.Run(":8087")
	if err != nil {
		return
	}
}
