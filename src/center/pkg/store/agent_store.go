// Package store provides Agent Store functionality
package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// StoreItemType defines store item types
type StoreItemType string

const (
	ItemTypeAgent   StoreItemType = "agent"
	ItemTypeSkill   StoreItemType = "skill"
	ItemTypeTemplate StoreItemType = "template"
	ItemTypeModel   StoreItemType = "model"
	ItemTypeDataset StoreItemType = "dataset"
)

// StoreItemStatus defines item status
type StoreItemStatus string

const (
	ItemPending   StoreItemStatus = "pending"
	ItemApproved  StoreItemStatus = "approved"
	ItemPublished StoreItemStatus = "published"
	ItemDeprecated StoreItemStatus = "deprecated"
	ItemRemoved   StoreItemStatus = "removed"
)

// StoreItem represents an item in the store
type StoreItem struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	DisplayName     string            `json:"display_name"`
	Type            StoreItemType     `json:"type"`
	Version         string            `json:"version"`
	Description     string            `json:"description"`
	LongDescription string            `json:"long_description"`
	Status          StoreItemStatus   `json:"status"`
	Category        string            `json:"category"`
	Tags            []string          `json:"tags"`
	Author          AuthorInfo        `json:"author"`
	License         string            `json:"license"`
	Homepage        string            `json:"homepage"`
	Repository      string            `json:"repository"`
	Documentation   string            `json:"documentation"`
	Icon            string            `json:"icon"`
	Screenshots     []string          `json:"screenshots"`
	Price           PriceInfo         `json:"price"`
	Rating          RatingInfo        `json:"rating"`
	Stats           ItemStats         `json:"stats"`
	Requirements    Requirements      `json:"requirements"`
	ConfigSchema    map[string]interface{} `json:"config_schema"`
	DefaultConfig   map[string]interface{} `json:"default_config"`
	Metadata        map[string]string `json:"metadata"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	PublishedAt     time.Time         `json:"published_at,omitempty"`
	Downloads       int64             `json:"downloads"`
	TenantID        string            `json:"tenant_id,omitempty"`
}

// AuthorInfo holds author information
type AuthorInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Company     string `json:"company,omitempty"`
	Website     string `json:"website,omitempty"`
	Verified    bool   `json:"verified"`
}

// PriceInfo holds pricing information
type PriceInfo struct {
	Model     string  `json:"model"` // free, paid, freemium, subscription
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Subscribe bool    `json:"subscribe"`
}

// RatingInfo holds rating information
type RatingInfo struct {
	Average   float64 `json:"average"`
	Count     int     `json:"count"`
	Stars5    int     `json:"stars_5"`
	Stars4    int     `json:"stars_4"`
	Stars3    int     `json:"stars_3"`
	Stars2    int     `json:"stars_2"`
	Stars1    int     `json:"stars_1"`
}

// ItemStats holds item statistics
type ItemStats struct {
	Downloads     int64   `json:"downloads"`
	Installs      int64   `json:"installs"`
	ActiveUsers   int64   `json:"active_users"`
	Favorites     int64   `json:"favorites"`
	Views         int64   `json:"views"`
	LastWeekDownloads int64 `json:"last_week_downloads"`
}

// Requirements holds item requirements
type Requirements struct {
	OFAVersion   string   `json:"ofa_version"`
	Platform     []string `json:"platform"`
	Dependencies []string `json:"dependencies"`
	MinCPU       int      `json:"min_cpu"`
	MinMemory    int      `json:"min_memory_mb"`
	GPU          bool     `json:"gpu"`
}

// StoreReview represents a review
type StoreReview struct {
	ID          string    `json:"id"`
	ItemID      string    `json:"item_id"`
	UserID      string    `json:"user_id"`
	UserName    string    `json:"user_name"`
	Rating      int       `json:"rating"` // 1-5
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Version     string    `json:"version"`
	Helpful     int       `json:"helpful"`
	Verified    bool      `json:"verified"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	TenantID    string    `json:"tenant_id,omitempty"`
}

// StoreCategory represents a category
type StoreCategory struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	Icon        string   `json:"icon"`
	ParentID    string   `json:"parent_id,omitempty"`
	ItemCount   int      `json:"item_count"`
	Featured    []string `json:"featured"`
}

// AgentStore manages the agent store
type AgentStore struct {
	storagePath string

	// Items
	items sync.Map // map[string]*StoreItem
	reviews sync.Map // map[string][]*StoreReview

	// Categories
	categories sync.Map // map[string]*StoreCategory

	// Search index
	searchIndex *SearchIndex

	// Statistics
	totalItems     int64
	totalDownloads int64

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// SearchIndex provides search functionality
type SearchIndex struct {
	index map[string][]string // term -> item IDs
	mu    sync.RWMutex
}

// NewAgentStore creates a new agent store
func NewAgentStore(storagePath string) (*AgentStore, error) {
	ctx, cancel := context.WithCancel(context.Background())

	os.MkdirAll(storagePath, 0755)

	store := &AgentStore{
		storagePath: storagePath,
		ctx:         ctx,
		cancel:      cancel,
		searchIndex: &SearchIndex{index: make(map[string][]string)},
	}

	// Initialize default categories
	store.initCategories()

	// Load existing data
	if err := store.load(); err != nil {
		log.Printf("Warning: failed to load store: %v", err)
	}

	// Build search index
	go store.buildSearchIndex()

	return store, nil
}

// initCategories initializes default categories
func (s *AgentStore) initCategories() {
	categories := []StoreCategory{
		{ID: "productivity", Name: "productivity", DisplayName: "Productivity", Description: "Boost your productivity", Icon: "⚡"},
		{ID: "communication", Name: "communication", DisplayName: "Communication", Description: "Communication tools", Icon: "💬"},
		{ID: "automation", Name: "automation", DisplayName: "Automation", Description: "Automate your workflows", Icon: "🤖"},
		{ID: "analytics", Name: "analytics", DisplayName: "Analytics", Description: "Data analytics agents", Icon: "📊"},
		{ID: "ai-ml", Name: "ai-ml", DisplayName: "AI & ML", Description: "Machine learning agents", Icon: "🧠"},
		{ID: "dev-tools", Name: "dev-tools", DisplayName: "Developer Tools", Description: "Development utilities", Icon: "🛠️"},
		{ID: "security", Name: "security", DisplayName: "Security", Description: "Security agents", Icon: "🔒"},
		{ID: "iot", Name: "iot", DisplayName: "IoT", Description: "IoT device management", Icon: "📡"},
	}

	for _, cat := range categories {
		s.categories.Store(cat.ID, &cat)
	}
}

// SubmitItem submits a new item to the store
func (s *AgentStore) SubmitItem(item *StoreItem) error {
	if item.ID == "" {
		item.ID = generateItemID(item.Type, item.Name)
	}

	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()
	item.Status = ItemPending
	item.Downloads = 0
	item.Stats = ItemStats{}
	item.Rating = RatingInfo{}

	// Validate
	if err := s.validateItem(item); err != nil {
		return err
	}

	s.items.Store(item.ID, item)
	s.totalItems++

	s.save()

	log.Printf("Store item submitted: %s (%s)", item.ID, item.Name)

	return nil
}

// validateItem validates a store item
func (s *AgentStore) validateItem(item *StoreItem) error {
	if item.Name == "" {
		return errors.New("item name required")
	}
	if item.Version == "" {
		return errors.New("item version required")
	}
	if item.Author.ID == "" {
		return errors.New("author ID required")
	}
	return nil
}

// ApproveItem approves a store item
func (s *AgentStore) ApproveItem(itemID string) error {
	item, err := s.GetItem(itemID)
	if err != nil {
		return err
	}

	item.Status = ItemApproved
	item.UpdatedAt = time.Now()

	s.save()

	log.Printf("Store item approved: %s", itemID)

	return nil
}

// PublishItem publishes a store item
func (s *AgentStore) PublishItem(itemID string) error {
	item, err := s.GetItem(itemID)
	if err != nil {
		return err
	}

	item.Status = ItemPublished
	item.PublishedAt = time.Now()
	item.UpdatedAt = time.Now()

	s.save()
	s.indexItem(item)

	log.Printf("Store item published: %s", itemID)

	return nil
}

// DeprecateItem deprecates a store item
func (s *AgentStore) DeprecateItem(itemID string) error {
	item, err := s.GetItem(itemID)
	if err != nil {
		return err
	}

	item.Status = ItemDeprecated
	item.UpdatedAt = time.Now()

	s.save()

	return nil
}

// GetItem retrieves a store item
func (s *AgentStore) GetItem(itemID string) (*StoreItem, error) {
	if v, ok := s.items.Load(itemID); ok {
		return v.(*StoreItem), nil
	}
	return nil, fmt.Errorf("item not found: %s", itemID)
}

// GetItemByName retrieves item by name
func (s *AgentStore) GetItemByName(name string) (*StoreItem, error) {
	var found *StoreItem
	s.items.Range(func(key, value interface{}) bool {
		item := value.(*StoreItem)
		if item.Name == name {
			found = item
			return false
		}
		return true
	})

	if found != nil {
		return found, nil
	}
	return nil, fmt.Errorf("item not found: %s", name)
}

// ListItems lists store items
func (s *AgentStore) ListItems(itemType StoreItemType, category string, limit int) []*StoreItem {
	var items []*StoreItem

	s.items.Range(func(key, value interface{}) bool {
		item := value.(*StoreItem)

		if item.Status != ItemPublished {
			return true
		}

		if itemType != "" && item.Type != itemType {
			return true
		}
		if category != "" && item.Category != category {
			return true
		}

		items = append(items, item)
		return true
	})

	// Sort by downloads
	sortItemsByDownloads(items)

	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}

	return items
}

// Search searches for items
func (s *AgentStore) Search(query string, filters SearchFilter) []*StoreItem {
	results := s.searchIndex.Search(query)

	var items []*StoreItem
	for _, id := range results {
		item, err := s.GetItem(id)
		if err != nil {
			continue
		}

		if !s.matchesFilter(item, filters) {
			continue
		}

		items = append(items, item)
	}

	return items
}

// SearchFilter defines search filters
type SearchFilter struct {
	Type     StoreItemType `json:"type,omitempty"`
	Category string        `json:"category,omitempty"`
	MinRating float64      `json:"min_rating,omitempty"`
	MaxPrice  float64      `json:"max_price,omitempty"`
	Tags     []string      `json:"tags,omitempty"`
}

// matchesFilter checks if item matches filter
func (s *AgentStore) matchesFilter(item *StoreItem, filter SearchFilter) bool {
	if filter.Type != "" && item.Type != filter.Type {
		return false
	}
	if filter.Category != "" && item.Category != filter.Category {
		return false
	}
	if filter.MinRating > 0 && item.Rating.Average < filter.MinRating {
		return false
	}
	if filter.MaxPrice > 0 && item.Price.Amount > filter.MaxPrice {
		return false
	}
	return true
}

// Download increments download count
func (s *AgentStore) Download(itemID string) error {
	item, err := s.GetItem(itemID)
	if err != nil {
		return err
	}

	s.mu.Lock()
	item.Downloads++
	item.Stats.Downloads++
	s.totalDownloads++
	s.mu.Unlock()

	s.save()

	return nil
}

// AddReview adds a review
func (s *AgentStore) AddReview(review *StoreReview) error {
	if review.ID == "" {
		review.ID = generateReviewID(review.ItemID, review.UserID)
	}

	review.CreatedAt = time.Now()
	review.UpdatedAt = time.Now()

	// Validate rating
	if review.Rating < 1 || review.Rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}

	// Store review
	reviews, _ := s.reviews.LoadOrStore(review.ItemID, []*StoreReview{})
	reviewList := reviews.([]*StoreReview)
	reviewList = append(reviewList, review)
	s.reviews.Store(review.ItemID, reviewList)

	// Update item rating
	s.updateItemRating(review.ItemID)

	s.save()

	return nil
}

// GetReviews retrieves reviews for an item
func (s *AgentStore) GetReviews(itemID string, limit int) []*StoreReview {
	if v, ok := s.reviews.Load(itemID); ok {
		reviews := v.([]*StoreReview)
		if limit > 0 && len(reviews) > limit {
			return reviews[len(reviews)-limit:]
		}
		return reviews
	}
	return nil
}

// updateItemRating updates item rating
func (s *AgentStore) updateItemRating(itemID string) {
	reviews := s.GetReviews(itemID, 0)
	if len(reviews) == 0 {
		return
	}

	item, err := s.GetItem(itemID)
	if err != nil {
		return
	}

	var total float64
	rating := RatingInfo{Count: len(reviews)}

	for _, r := range reviews {
		total += float64(r.Rating)
		switch r.Rating {
		case 5:
			rating.Stars5++
		case 4:
			rating.Stars4++
		case 3:
			rating.Stars3++
		case 2:
			rating.Stars2++
		case 1:
			rating.Stars1++
		}
	}

	rating.Average = total / float64(len(reviews))
	item.Rating = rating
}

// GetCategories retrieves categories
func (s *AgentStore) GetCategories() []*StoreCategory {
	var categories []*StoreCategory
	s.categories.Range(func(key, value interface{}) bool {
		cat := value.(*StoreCategory)
		// Update item count
		cat.ItemCount = len(s.ListItems("", cat.ID, 0))
		categories = append(categories, cat)
		return true
	})
	return categories
}

// GetFeatured retrieves featured items
func (s *AgentStore) GetFeatured(itemType StoreItemType, limit int) []*StoreItem {
	items := s.ListItems(itemType, "", 100)

	// Sort by rating and downloads
	sortItemsByRating(items)

	if len(items) > limit {
		items = items[:limit]
	}

	return items
}

// GetStats returns store statistics
func (s *AgentStore) GetStats() map[string]interface{} {
	typeCount := make(map[StoreItemType]int64)

	s.items.Range(func(key, value interface{}) bool {
		item := value.(*StoreItem)
		if item.Status == ItemPublished {
			typeCount[item.Type]++
		}
		return true
	})

	return map[string]interface{}{
		"total_items":     s.totalItems,
		"total_downloads": s.totalDownloads,
		"items_by_type":   typeCount,
		"categories":      len(s.GetCategories()),
	}
}

// indexItem adds item to search index
func (s *AgentStore) indexItem(item *StoreItem) {
	s.searchIndex.mu.Lock()
	defer s.searchIndex.mu.Unlock()

	// Index name
	for _, term := range tokenize(item.Name) {
		s.searchIndex.index[term] = appendUnique(s.searchIndex.index[term], item.ID)
	}

	// Index description
	for _, term := range tokenize(item.Description) {
		s.searchIndex.index[term] = appendUnique(s.searchIndex.index[term], item.ID)
	}

	// Index tags
	for _, tag := range item.Tags {
		for _, term := range tokenize(tag) {
			s.searchIndex.index[term] = appendUnique(s.searchIndex.index[term], item.ID)
		}
	}
}

// buildSearchIndex builds search index
func (s *AgentStore) buildSearchIndex() {
	s.items.Range(func(key, value interface{}) bool {
		item := value.(*StoreItem)
		if item.Status == ItemPublished {
			s.indexItem(item)
		}
		return true
	})
}

// load loads store from disk
func (s *AgentStore) load() error {
	data, err := os.ReadFile(filepath.Join(s.storagePath, "store.json"))
	if err != nil {
		return err
	}

	var items []*StoreItem
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}

	for _, item := range items {
		s.items.Store(item.ID, item)
	}

	return nil
}

// save saves store to disk
func (s *AgentStore) save() error {
	var items []*StoreItem
	s.items.Range(func(key, value interface{}) bool {
		items = append(items, value.(*StoreItem))
		return true
	})

	data, err := json.Marshal(items)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(s.storagePath, "store.json"), data, 0644)
}

// Close closes the store
func (s *AgentStore) Close() {
	s.cancel()
	s.save()
}

// Helper functions

func generateItemID(itemType StoreItemType, name string) string {
	return fmt.Sprintf("%s-%s-%d", itemType, name, time.Now().UnixNano())
}

func generateReviewID(itemID, userID string) string {
	return fmt.Sprintf("review-%s-%s-%d", itemID, userID, time.Now().UnixNano())
}

func tokenize(s string) []string {
	// Simple tokenization
	var tokens []string
	word := ""
	for _, c := range s {
		if c == ' ' || c == '-' || c == '_' {
			if len(word) > 1 {
				tokens = append(tokens, word)
			}
			word = ""
		} else {
			word += string(c)
		}
	}
	if len(word) > 1 {
		tokens = append(tokens, word)
	}
	return tokens
}

func appendUnique(slice []string, s string) []string {
	for _, v := range slice {
		if v == s {
			return slice
		}
	}
	return append(slice, s)
}

func sortItemsByDownloads(items []*StoreItem) {
	// Simple bubble sort
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].Downloads > items[i].Downloads {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

func sortItemsByRating(items []*StoreItem) {
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].Rating.Average > items[i].Rating.Average {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

// Search method for SearchIndex
func (si *SearchIndex) Search(query string) []string {
	si.mu.RLock()
	defer si.mu.RUnlock()

	var results []string
	for _, term := range tokenize(query) {
		if ids, ok := si.index[term]; ok {
			results = append(results, ids...)
		}
	}
	return results
}