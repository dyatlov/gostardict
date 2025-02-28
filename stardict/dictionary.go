package stardict

import (
	"bytes"
	"encoding/binary"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

// Translation contains translation items
type Translation struct {
	Parts []*TranslationItem
}

// TranslationItem contain single translation item
type TranslationItem struct {
	Type rune
	Data []byte
}

// Dictionary stardict dictionary
type Dictionary struct {
	dict *Dict
	idx  *Idx
	info *Info
}

// Translate translates given item
func (d Dictionary) Translate(item string) (items []*Translation) {
	senses := d.idx.Get(item)

	for _, seq := range senses {
		sense := d.dict.GetSequence(seq.Offset, seq.Size)

		var transItems []*TranslationItem

		if _, ok := d.info.Options["sametypesequence"]; ok {
			transItems = d.translateWithSametypesequence(sense)
		} else {
			transItems = d.translateWithoutSametypesequence(sense)
		}

		items = append(items, &Translation{Parts: transItems})
	}

	return
}

func (d Dictionary) translateWithSametypesequence(data []byte) (items []*TranslationItem) {
	seq := d.info.Options["sametypesequence"]

	seqLen := len(seq)

	var dataPos int
	dataSize := len(data)

	for i, t := range seq {
		switch t {
		case 'm', 'l', 'g', 't', 'x', 'y', 'k', 'w', 'h', 'r':
			// if last seq item
			if i == seqLen-1 {
				items = append(items, &TranslationItem{Type: t, Data: data[dataPos:dataSize]})
			} else {
				end := bytes.IndexRune(data[dataPos:], '\000')
				items = append(items, &TranslationItem{Type: t, Data: data[dataPos : dataPos+end+1]})
				dataPos += end + 1
			}
		case 'W', 'P':
			if i == seqLen-1 {
				items = append(items, &TranslationItem{Type: t, Data: data[dataPos:dataSize]})
			} else {
				size := binary.BigEndian.Uint32(data[dataPos : dataPos+4])
				items = append(items, &TranslationItem{Type: t, Data: data[dataPos+4 : dataPos+int(size)+5]})
				dataPos += int(size) + 5
			}
		}
	}

	return
}

func (d Dictionary) translateWithoutSametypesequence(data []byte) (items []*TranslationItem) {
	var dataPos int
	dataSize := len(data)

	for {
		t := data[dataPos]

		dataPos++

		switch t {
		case 'm', 'l', 'g', 't', 'x', 'y', 'k', 'w', 'h', 'r':
			end := bytes.IndexRune(data[dataPos:], '\000')

			if end < 0 { // last item
				items = append(items, &TranslationItem{Type: rune(t), Data: data[dataPos:dataSize]})
				dataPos = dataSize
			} else {
				items = append(items, &TranslationItem{Type: rune(t), Data: data[dataPos : dataPos+end+1]})
				dataPos += end + 1
			}
		case 'W', 'P':
			size := binary.BigEndian.Uint32(data[dataPos : dataPos+4])
			items = append(items, &TranslationItem{Type: rune(t), Data: data[dataPos+4 : dataPos+int(size)+5]})
			dataPos += int(size) + 5
		}

		if dataPos >= dataSize {
			break
		}
	}

	return
}

// GetBookName returns book name
func (d Dictionary) GetBookName() string {
	return d.info.Options["bookname"]
}

// GetWordCount returns number of words in the dictionary
func (d Dictionary) GetWordCount() uint64 {
	num, _ := strconv.ParseUint(d.info.Options["wordcount"], 10, 64)

	return num
}

// NewDictionary returns a new Dictionary
// path - path to dictionary files
// name - name of dictionary to parse
func NewDictionary(path string, name string) (*Dictionary, error) {
	d := new(Dictionary)

	path = filepath.Clean(path)

	dictDzPath := filepath.Join(path, name+".dict.dz")
	dictPath := filepath.Join(path, name+".dict")

	idxPath := filepath.Join(path, name+".idx")
	infoPath := filepath.Join(path, name+".ifo")

	if _, err := os.Stat(idxPath); err != nil {
		return nil, err
	}

	// we should have either .dict.dz or .dict file
	if _, err := os.Stat(dictDzPath); os.IsNotExist(err) {
		if _, err := os.Stat(dictPath); os.IsNotExist(err) {
			return nil, err
		}
	} else {
		dictPath = dictDzPath
	}

	info, err := ReadInfo(infoPath)

	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		log.Printf("ifo file not found, using defaults")
		info = &Info{
			Version: "3.0.0",
			Options: make(map[string]string),
		}
	}

	idx, err := ReadIndex(idxPath, info)

	if err != nil {
		return nil, err
	}

	dict, err := ReadDict(dictPath, info)

	if err != nil {
		return nil, err
	}

	d.info = info
	d.idx = idx
	d.dict = dict

	return d, nil
}
