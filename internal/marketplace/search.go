package marketplace

import (
	"context"
	"slices"
	"strings"
)

// Searcher provides plugin search functionality.
type Searcher struct {
	client *Client
}

// NewSearcher creates a new plugin searcher.
func NewSearcher(client *Client) *Searcher {
	return &Searcher{
		client: client,
	}
}

// Search searches for plugins with given query and options.
func (s *Searcher) Search(ctx context.Context, query string, opts SearchOptions) ([]PluginInfo, error) {
	plugins, err := s.client.Search(ctx, query, opts)
	if err != nil {
		return nil, err
	}

	if opts.SortBy != "" {
		plugins = s.sortPlugins(plugins, opts.SortBy)
	}

	if opts.Category != "" {
		plugins = s.filterByCategory(plugins, opts.Category)
	}

	if opts.Author != "" {
		plugins = s.filterByAuthor(plugins, opts.Author)
	}

	return plugins, nil
}

func (s *Searcher) sortPlugins(plugins []PluginInfo, sortBy string) []PluginInfo {
	result := make([]PluginInfo, len(plugins))
	copy(result, plugins)

	switch sortBy {
	case "downloads":
		slices.SortFunc(result, func(a, b PluginInfo) int {
			return b.Downloads - a.Downloads
		})
	case "rating":
		slices.SortFunc(result, func(a, b PluginInfo) int {
			if a.Rating > b.Rating {
				return -1
			}
			if a.Rating < b.Rating {
				return 1
			}
			return 0
		})
	case "name":
		slices.SortFunc(result, func(a, b PluginInfo) int {
			return strings.Compare(a.Name, b.Name)
		})
	}

	return result
}

func (s *Searcher) filterByCategory(plugins []PluginInfo, category string) []PluginInfo {
	result := []PluginInfo{}
	for _, p := range plugins {
		if strings.EqualFold(p.Category, category) {
			result = append(result, p)
		}
	}
	return result
}

func (s *Searcher) filterByAuthor(plugins []PluginInfo, author string) []PluginInfo {
	result := []PluginInfo{}
	for _, p := range plugins {
		if strings.EqualFold(p.Author, author) {
			result = append(result, p)
		}
	}
	return result
}

// SearchByTags searches for plugins by tags.
func (s *Searcher) SearchByTags(ctx context.Context, tags []string) ([]PluginInfo, error) {
	if len(tags) == 0 {
		return []PluginInfo{}, nil
	}

	query := strings.Join(tags, ",")

	plugins, err := s.client.Search(ctx, query, SearchOptions{})
	if err != nil {
		return nil, err
	}

	result := []PluginInfo{}
	for _, p := range plugins {
		if s.hasAllTags(p.Tags, tags) {
			result = append(result, p)
		}
	}

	return result, nil
}

// GetCategories returns available plugin categories.
func (s *Searcher) GetCategories(_ context.Context) ([]string, error) {
	return []string{
		"skills",
		"agents",
		"rules",
		"development",
		"testing",
		"devops",
		"enterprise",
		"utilities",
	}, nil
}

func (s *Searcher) hasAllTags(pluginTags []string, searchTags []string) bool {
	for _, searchTag := range searchTags {
		found := false
		for _, pluginTag := range pluginTags {
			if strings.EqualFold(pluginTag, searchTag) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// GetPopularPlugins returns most downloaded plugins.
func (s *Searcher) GetPopularPlugins(ctx context.Context, limit int) ([]PluginInfo, error) {
	return s.Search(ctx, "", SearchOptions{
		SortBy: "downloads",
		Limit:  limit,
	})
}

// GetTopRatedPlugins returns highest rated plugins.
func (s *Searcher) GetTopRatedPlugins(ctx context.Context, limit int) ([]PluginInfo, error) {
	return s.Search(ctx, "", SearchOptions{
		SortBy: "rating",
		Limit:  limit,
	})
}

// GetPluginsByAuthor returns plugins by specific author.
func (s *Searcher) GetPluginsByAuthor(ctx context.Context, author string) ([]PluginInfo, error) {
	plugins, err := s.client.Search(ctx, "", SearchOptions{})
	if err != nil {
		return nil, err
	}

	return s.filterByAuthor(plugins, author), nil
}
