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
	"time"
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
	m := &Media{}
	m.Categories = make(map[uint8]MediaCategory)
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
