package media

import (
	"encoding/json"
	"errors"
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
	Categories []*Category
	IDs        []uint8
	IDsSum     uint8
}

func New() (*Media, error) {
	rand.Seed(time.Now().Unix())
	folder := filepath.Join(".", "assets")
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return nil, fmt.Errorf("mediaCategories read dir: %w", err)
	}
	m := &Media{Categories: make([]*Category, 0)}
	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".json" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(folder, f.Name()))
		if err != nil {
			log.Println(err)
			continue
		}
		var mc Category
		if err := json.Unmarshal(b, &mc); err != nil {
			log.Println(err)
			continue
		}
		m.Categories = append(m.Categories, &mc)
		m.IDsSum = m.IDsSum | mc.ID
	}

	if len(m.Categories) == 0 {
		return nil, errors.New("no media categories")
	}

	sort.Slice(m.Categories, func(i, j int) bool {
		return m.Categories[i].ID < m.Categories[j].ID
	})

	m.IDs = make([]uint8, 0, len(m.Categories))
	for _, mc := range m.Categories {
		m.IDs = append(m.IDs, mc.ID)
		//fmt.Println(*mc)
	}

	return m, nil
}

type Category struct {
	Name     string   `json:"name"`
	ID       uint8    `json:"id"`
	Internal bool     `json:"internal"`
	Path     string   `json:"path,omitempty"`
	Files    []string `json:"files"`

	TxtName string `json:"txt_name"` // _ - space, 18p - 18+
}

func (m *Media) ExtractCatIDs(catJoinedIDs uint8) []uint8 {
	if catJoinedIDs == 0 || catJoinedIDs == m.IDsSum {
		return m.IDs
	}
	cats := make([]uint8, 0)
	for _, id := range m.IDs {
		if id&catJoinedIDs > 0 {
			cats = append(cats, id)
		}
	}
	return cats
}

// RandomImage 1 && 2 && 4 : can be 1, 2, 3, 4, 5, 6, 7  // but 0 - all
func (m *Media) RandomImage(catJoinedIDs uint8) string {
	fromCats := m.ExtractCatIDs(catJoinedIDs)
	n := rand.Intn(len(fromCats))

	cat := m.GetCatById(fromCats[n])
	if cat == nil {
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

func (m *Media) GetCatById(catID uint8) *Category {
	for key := range m.Categories {
		if m.Categories[key].ID == catID {
			return m.Categories[key]
		}
	}
	return nil
}
