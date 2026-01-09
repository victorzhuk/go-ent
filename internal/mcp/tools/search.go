package tools

import (
	"math"
	"sort"
	"strings"
	"unicode"
)

// SearchIndex provides TF-IDF based search functionality for tool discovery.
type SearchIndex struct {
	docs      []Document
	termFreqs map[string]map[int]float64 // term -> doc_id -> freq
	docFreqs  map[string]int             // term -> num_docs containing
	totalDocs int
}

// Document represents a searchable document (tool).
type Document struct {
	ID       int
	ToolName string
	Terms    []string
	TF       map[string]float64
}

// SearchResult represents a search result with relevance score.
type SearchResult struct {
	ToolName string
	Score    float64
	DocID    int
}

// NewSearchIndex creates a new TF-IDF search index.
func NewSearchIndex() *SearchIndex {
	return &SearchIndex{
		docs:      make([]Document, 0),
		termFreqs: make(map[string]map[int]float64),
		docFreqs:  make(map[string]int),
	}
}

// Index builds the TF-IDF index from documents.
func (s *SearchIndex) Index(docs []Document) error {
	s.docs = docs
	s.totalDocs = len(docs)

	// Calculate term frequencies and document frequencies
	for _, doc := range docs {
		termCounts := make(map[string]int)
		for _, term := range doc.Terms {
			termCounts[term]++
		}

		// Calculate TF for this document
		totalTerms := len(doc.Terms)
		for term, count := range termCounts {
			tf := float64(count) / float64(totalTerms)

			// Store TF
			if s.termFreqs[term] == nil {
				s.termFreqs[term] = make(map[int]float64)
			}
			s.termFreqs[term][doc.ID] = tf

			// Count document frequency
			s.docFreqs[term]++
		}
	}

	return nil
}

// Search performs TF-IDF search and returns ranked results.
func (s *SearchIndex) Search(query string, limit int) []SearchResult {
	queryTerms := extractTerms(query)
	if len(queryTerms) == 0 {
		return nil
	}

	// Calculate scores for each document
	scores := make(map[int]float64)
	for _, doc := range s.docs {
		score := s.calculateScore(doc, queryTerms)
		if score > 0 {
			scores[doc.ID] = score
		}
	}

	// Convert to sorted results
	results := make([]SearchResult, 0, len(scores))
	for docID, score := range scores {
		results = append(results, SearchResult{
			ToolName: s.docs[docID].ToolName,
			Score:    score,
			DocID:    docID,
		})
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Limit results
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results
}

// calculateScore calculates TF-IDF score for a document given query terms.
func (s *SearchIndex) calculateScore(doc Document, queryTerms []string) float64 {
	score := 0.0

	for _, term := range queryTerms {
		// Get TF for this term in this document
		tf, hasTerm := s.termFreqs[term][doc.ID]
		if !hasTerm {
			continue
		}

		// Calculate IDF
		docFreq := s.docFreqs[term]
		if docFreq == 0 {
			continue
		}

		idf := math.Log(float64(s.totalDocs) / float64(docFreq))

		// TF-IDF score
		score += tf * idf
	}

	return score
}

// extractTerms tokenizes and normalizes text into searchable terms.
func extractTerms(text string) []string {
	text = strings.ToLower(text)

	// Split on non-alphanumeric characters
	var terms []string
	var current strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			current.WriteRune(r)
		} else if current.Len() > 0 {
			term := current.String()
			if !isStopword(term) && len(term) > 1 {
				terms = append(terms, term)
			}
			current.Reset()
		}
	}

	// Handle last term
	if current.Len() > 0 {
		term := current.String()
		if !isStopword(term) && len(term) > 1 {
			terms = append(terms, term)
		}
	}

	return terms
}

// isStopword checks if a term is a common stopword.
func isStopword(term string) bool {
	stopwords := map[string]bool{
		"a": true, "an": true, "and": true, "are": true, "as": true,
		"at": true, "be": true, "by": true, "for": true, "from": true,
		"has": true, "he": true, "in": true, "is": true, "it": true,
		"its": true, "of": true, "on": true, "or": true, "that": true,
		"the": true, "to": true, "was": true, "will": true, "with": true,
	}
	return stopwords[term]
}

// BuildDocument creates a searchable document from tool metadata.
func BuildDocument(id int, toolName, description string) Document {
	text := toolName + " " + description
	terms := extractTerms(text)

	// Calculate TF
	termCounts := make(map[string]int)
	for _, term := range terms {
		termCounts[term]++
	}

	tf := make(map[string]float64)
	totalTerms := len(terms)
	for term, count := range termCounts {
		tf[term] = float64(count) / float64(totalTerms)
	}

	return Document{
		ID:       id,
		ToolName: toolName,
		Terms:    terms,
		TF:       tf,
	}
}
