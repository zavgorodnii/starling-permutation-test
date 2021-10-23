package src

import (
	"log"
	"regexp"
	"strings"
	"unicode/utf8"
	"github.com/pkg/errors"
	"github.com/tealeg/xlsx"
)

const (
	swadeshIDCol   = 0
	swadeshWordCol = 1
	groupsStartCol = 2
)

var (
	Laryngeals        string
	VowelsAndFeatures string
	Glides            string
	LabialGlides      string
	IPAS							string
)

type SoundClassesDecoder struct {
	SoundToClassID map[rune]string
}

func NewSoundClassesDecoder(classesPath string) (*SoundClassesDecoder, error) {
	out := &SoundClassesDecoder{
		SoundToClassID: map[rune]string{},
	}

	classesFile, err := xlsx.OpenFile(classesPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read %s", classesPath)
	}

	if len(classesFile.Sheets) != 1 {
		return nil, errors.New("single sheet is expected")
	}

	for idx, row := range classesFile.Sheets[0].Rows {
		if len(row.Cells) < 2 {
			return nil, errors.Errorf("row %d has less than 2 cells", idx)
		}
		var (
			classMembers = strings.TrimSpace(row.Cells[0].String())
			className    = strings.TrimSpace(row.Cells[1].String())
			classID      = classMembers[:1]
		)
		IPAS = strings.Join([]string{IPAS,classMembers}, "")
		switch className {
		case "Laryngeals":
			Laryngeals = classID
		case "Vowels and features":
			VowelsAndFeatures = classID
		case "Glides":
			Glides = classID
		case "Labial glides":
			LabialGlides = classID
		}

		for _, classMember := range classMembers {
			out.SoundToClassID[classMember] = classID
		}
	}

	if len(Laryngeals) < 0 {
		log.Println("WARNING: Laryngeals not found, this might lead to incorrect behavior")
	}
	if len(VowelsAndFeatures) < 0 {
		log.Println("WARNING: VowelsAndFeatures not found, this might lead to incorrect behavior")
	}
	if len(Glides) < 0 {
		log.Println("WARNING: Glides not found, this might lead to incorrect behavior")
	}
	if len(LabialGlides) < 0 {
		log.Println("WARNING: LabialGlides not found, this might lead to incorrect behavior")
	}

	return out, nil
}

func (d *SoundClassesDecoder) Decode(listsPath string, selected map[string]bool) ([]*Wordlist, error) {
	groupToWordlist := map[string]*Wordlist{}

	wordlistsFile, err := xlsx.OpenFile(listsPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read %s", listsPath)
	}

	//if len(wordlistsFile.Sheets) != 1 {
	//	return nil, errors.New("single sheet is expected")
	//} Deleted by D. Krylov, 2021

	if len(wordlistsFile.Sheets[0].Rows) < 2 {
		return nil, errors.New("document is malformed: less than 2 rows is present")
	}

	var (
		sortedGroupNames []string
		headerRow        = wordlistsFile.Sheets[0].Rows[0].Cells
		allSelected      bool
	)
	if selected == nil {
		selected = map[string]bool{}
		allSelected = true
	}

	var maxGroupIdx = groupsStartCol
	for groupIdx := groupsStartCol; groupIdx < len(headerRow); groupIdx++ {
		var groupName = headerRow[groupIdx].String()
		if len(groupName) > 0 {
			maxGroupIdx++
		}

		if strings.HasSuffix(groupName, "NUM") {
			continue
		}

		if _, ok := selected[groupName]; allSelected || ok {
			selected[groupName] = true
			sortedGroupNames = append(sortedGroupNames, groupName)
			groupToWordlist[groupName] = &Wordlist{Group: groupName}
		}
	}

	var lastSwadeshID = 0
	for idx := 1; idx < len(wordlistsFile.Sheets[0].Rows); idx++ {
		if strings.HasPrefix(wordlistsFile.Sheets[0].Rows[idx].Cells[0].String(), "0"){
			continue
		}
		row := wordlistsFile.Sheets[0].Rows[idx].Cells
		swadeshID, err := row[swadeshIDCol].Int()
		if err != nil {
			return nil, errors.Wrapf(err, "row %d, column %d", idx, swadeshIDCol)
		}

		var swadeshWord = strings.TrimSpace(row[swadeshWordCol].String())
		for groupIdx := groupsStartCol; groupIdx < maxGroupIdx; groupIdx++ {
			if _, ok := selected[headerRow[groupIdx].String()]; !ok {
				continue
			}
			// Some group columns are followed by a column containing cognition indices.
			var (
				skipColumn bool
				ignoreForm bool
			)
			if groupIdx+1 < maxGroupIdx {
				if maybeCognitiveIndex, err := row[groupIdx+1].Int(); err == nil {
					skipColumn = true
					if maybeCognitiveIndex < 0 {
						ignoreForm = true
					}
				}
			}

			var (
				groupName          = headerRow[groupIdx].String()
				swadeshWordCleaner = regexp.MustCompile("[0-9]|\\[.*\\]")
			)
			// Start a new Swadesh word.
			if lastSwadeshID != swadeshID {
				word := &Word{
					Group:       groupName,
					SwadeshID:   swadeshID,
					SwadeshWord: swadeshWordCleaner.ReplaceAllString(swadeshWord, ""),
				}
				groupToWordlist[groupName].List = append(groupToWordlist[groupName].List, word)
			}

			var form = strings.TrimSpace(row[groupIdx].String())
			if len(form) < 1 {
				if skipColumn {
					groupIdx++
				}
				continue
			}

			var clean, decoded, broomed, withoutbrackets = d.decodeForm(form)
			if !ignoreForm {
				lastWord := groupToWordlist[groupName].List[len(groupToWordlist[groupName].List)-1]
				lastWord.Forms = append(lastWord.Forms, form)
				lastWord.CleanForms = append(lastWord.CleanForms, clean...)
				lastWord.DecodedForms = append(lastWord.DecodedForms, decoded...)
				lastWord.BroomedSymbols = append(lastWord.BroomedSymbols, broomed...)
				lastWord.WithoutBrackets = append(lastWord.WithoutBrackets, withoutbrackets...)
			}

			if skipColumn {
				groupIdx++
			}
		}

		lastSwadeshID = swadeshID
	}

	var out []*Wordlist
	for _, groupName := range sortedGroupNames {
		if len(groupToWordlist[groupName].List) > 0 {
			out = append(out, groupToWordlist[groupName])
		}
	}

	return out, nil
}

func (d *SoundClassesDecoder) decodeForm(form string) (clean []string, decoded []string, broomed []string, withoutbrackets []string) {
	r, _ := regexp.Compile("\\{.*?\\}|\\(.*?\\)")
	form = r.ReplaceAllString(form, "")
	b, _ := regexp.Compile("[^ !\\-,=/#~"+IPAS+"]")
	broomed = b.FindAllString(form, -1)

	form = strings.Replace(form, "*", "", -1)

	if strings.Contains(form, "~") {
		withoutbrackets = append(withoutbrackets, strings.Split(form, "~")...)
		form = b.ReplaceAllString(form, "")
		clean = append(clean, strings.Split(form, "~")...)
	} else if strings.Contains(form, "/") {
		withoutbrackets = append(withoutbrackets, strings.Split(form, "/")...)
		form = b.ReplaceAllString(form, "")
		clean = append(clean, strings.Split(form, "/")...)
	} else {
		withoutbrackets = append(withoutbrackets, form)
		form = b.ReplaceAllString(form, "")
		clean = append(clean, form)
	}

	for idx, form := range clean {
		clean[idx] = d.cleanseForm(form)
	}

	decoded = make([]string, len(clean))
	for idx, word := range clean {
		var (
			decodedForm string
			charIdx     int
		)
		for _, char := range word {
			if classID, ok := d.SoundToClassID[char]; ok {
				switch classID {
				case Laryngeals, VowelsAndFeatures:
					if len(decodedForm) < 2 && (charIdx == 0 || charIdx >= utf8.RuneCountInString(word)-1) {
						decodedForm += Laryngeals
					}
				case Glides, LabialGlides:
					if len(decodedForm) < 2 {
						if charIdx == 0 {
							decodedForm += classID
						}
						if charIdx >= utf8.RuneCountInString(word)-1 {
							decodedForm += Laryngeals
						}
					}
				default:
					decodedForm += classID
				}
			}

			charIdx++
		}

		if len(decodedForm) == 1 {
			decodedForm += VowelsAndFeatures
		}

		decoded[idx] = decodedForm

	}

	return clean, decoded, broomed, withoutbrackets
}

func (d SoundClassesDecoder) cleanseForm(form string) (out string) {
	form = strings.TrimSpace(form)
	for _, char := range form {
		switch char {
		case '=':
			out = ""
		case '-':
			return strings.TrimSpace(out)
		case ' ':
			return strings.TrimSpace(out)
		default:
			out += string(char)
		}
	}

	return strings.TrimSpace(out)
}
