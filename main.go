package main

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

// Item 示例资源
type Item struct {
	ID    int     `json:"id"`
	Name  string  `json:"name" binding:"required"`
	Price float64 `json:"price" binding:"required"`
}

// Store 线程安全的内存存储
type Store struct {
	mu    sync.Mutex
	items map[int]Item
	next  int
}

func NewStore() *Store {
	return &Store{
		items: map[int]Item{
			1: {ID: 1, Name: "apple", Price: 3.5},
			2: {ID: 2, Name: "banana", Price: 2.0},
		},
		next: 3,
	}
}

// List 查询全部
func (s *Store) List() []Item {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := make([]Item, 0, len(s.items))
	for _, it := range s.items {
		items = append(items, it)
	}
	return items
}

// Get 按 ID 查询
func (s *Store) Get(id int) (Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	it, ok := s.items[id]
	return it, ok
}

// Create 新增,ID 自动分配
func (s *Store) Create(item Item) Item {
	s.mu.Lock()
	defer s.mu.Unlock()
	item.ID = s.next
	s.next++
	s.items[item.ID] = item
	return item
}

// Update 按 ID 更新(不存在返回 false)
func (s *Store) Update(id int, item Item) (Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return Item{}, false
	}
	item.ID = id
	s.items[id] = item
	return item, true
}

// Delete 按 ID 删除(不存在返回 false)
func (s *Store) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return false
	}
	delete(s.items, id)
	return true
}

func main() {
	store := NewStore()
	r := gin.Default()

	// GET /items    查询全部
	r.GET("/items", func(c *gin.Context) {
		c.JSON(http.StatusOK, store.List())
	})

	// GET /items/:id 查询单个
	r.GET("/items/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		item, ok := store.Get(id)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
			return
		}
		c.JSON(http.StatusOK, item)
	})

	// POST /items    新增
	r.POST("/items", func(c *gin.Context) {
		var item Item
		if err := c.ShouldBindJSON(&item); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, store.Create(item))
	})

	// PUT /items/:id 更新
	r.PUT("/items/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		var item Item
		if err := c.ShouldBindJSON(&item); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		updated, ok := store.Update(id, item)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
			return
		}
		c.JSON(http.StatusOK, updated)
	})

	// DELETE /items/:id 删除
	r.DELETE("/items/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		if !store.Delete(id) {
			c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	r.Run(":8080")
}
