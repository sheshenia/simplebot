package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	DefaultImageHref = "https://telegram.org/img/t_logo.png"
)

type Media struct {
	Categories map[uint8]MediaCategory
	IDs        UIntSlice
	IDsSum     uint8
}

func NewMedia() (*Media, error) {
	rand.Seed(time.Now().Unix())
	folder := filepath.Join(".", "assets")
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return nil, fmt.Errorf("mediaCategories read dir: %w", err)
	}
	m := &Media{Categories: make(map[uint8]MediaCategory)}
	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".json" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(folder, f.Name()))
		if err != nil {
			log.Println(err)
			continue
		}
		var mf MediaCategory
		if err := json.Unmarshal(b, &mf); err != nil {
			log.Println(err)
			continue
		}
		m.Categories[mf.ID] = mf
	}

	m.IDs = make([]uint8, 0, len(m.Categories))

	for _, cat := range m.Categories {
		m.IDs = append(m.IDs, cat.ID)
		m.IDsSum = m.IDsSum | cat.ID
	}
	sort.Sort(m.IDs)
	return m, nil
}

type MediaCategory struct {
	Name     string   `json:"name"`
	ID       uint8    `json:"id"`
	Internal bool     `json:"internal"`
	Path     string   `json:"path,omitempty"`
	Files    []string `json:"files"`

	TxtName string `json:"txt_name"` // _ - space, 18p - 18+
}

type UIntSlice []uint8

func (x UIntSlice) Len() int           { return len(x) }
func (x UIntSlice) Less(i, j int) bool { return x[i] < x[j] }
func (x UIntSlice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

func (m *Media) extractCatIDs(catJoinedIDs uint8) []uint8 {
	if catJoinedIDs == 0 || catJoinedIDs == m.IDsSum {
		return m.IDs
	}
	cats := []uint8{}
	for _, id := range m.IDs {
		if id&catJoinedIDs > 0 {
			cats = append(cats, id)
		}
	}
	return cats
}

// randomImage 1 && 2 && 4 : can be 1, 2, 3, 4, 5, 6, 7  // but 0 - all
func (m *Media) randomImage(catJoinedIDs uint8) string {
	fromCats := m.extractCatIDs(catJoinedIDs)
	n := rand.Intn(len(fromCats))

	cat, ok := m.Categories[fromCats[n]]
	if !ok {
		log.Println("!ok MediaCategories[fromNames[catID]]", fromCats[n])
		return DefaultImageHref
	}
	if len(cat.Files) == 0 {
		return DefaultImageHref
	}
	fileName := cat.Files[rand.Intn(len(cat.Files))]
	if strings.HasPrefix(fileName, "http") {
		return fileName
	}
	return cat.Path + fileName
}
