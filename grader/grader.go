package grader

import (
	"fmt"
	"github.com/jrabasco/quiz-grader/files"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Kind int

const (
	MC Kind = iota
	FREE
)

type Answer interface {
	Grade(string, string, int, int) int
}

type Mult struct {
	answer int
}

func (mc Mult) normalise(sub string) (int, bool) {
	tr := strings.TrimSpace(sub)
	tr = strings.ToLower(tr)
	if len(tr) == 0 {
		return 0, false
	}

	flag := len(tr) != 1

	ans, err := strconv.Atoi(tr)
	if err == nil {
		return ans, flag
	}

	// 1-indexed
	return int(tr[0]-'a') + 1, flag
}

func (mc Mult) Grade(sub string, player string, section int, question int) int {
	norm, flag := mc.normalise(sub)
	if flag {
		fmt.Printf("FLAG! %s in section %d question %d\n", player, section, question)
	}
	if norm == mc.answer {
		return 1
	}
	return 0
}

func parseMC(ansStr string) (*Mult, error) {
	answer, err := strconv.Atoi(ansStr)
	return &Mult{answer}, err
}

type Free struct {
	poss   []string
	points int
}

func (fr Free) normalise(sub string) string {
	tr := strings.TrimSpace(sub)
	tr = strings.ToLower(tr)
	return tr
}

func (fr Free) Grade(sub string, player string, section int, question int) int {
	norm := fr.normalise(sub)
	for _, p := range fr.poss {
		if p == norm {
			return fr.points
		}
	}

	bestStr := strings.Join(fr.poss, ", ")

	fmt.Printf("%s in section %d question %d replied '%s', best=[%s], points? ", player, section, question, norm, bestStr)
	pts := -1
	for pts < 0 {
		var judge string
		nb, err := fmt.Scanln(&judge)
		if err != nil || nb != 1 {
			pts = 0
			break
		}
		pts, err = strconv.Atoi(judge)
		if err != nil {
			pts = -1
		}
	}
	return pts
}
func parseFree(ansStr string, ptsStr string) (*Free, error) {
	var f Free
	pts, err := strconv.Atoi(ptsStr)
	if err != nil {
		return &f, err
	}
	f.points = pts
	f.poss = strings.Split(ansStr, ",")
	if len(f.poss) < 1 {
		return &f, fmt.Errorf("empty possibilities")
	}
	return &f, nil
}

func parseAnswer(line string) (Answer, error) {
	parts := strings.Split(line, ":")
	if len(parts) < 2 {
		return nil, fmt.Errorf("malformed answer: %s", line)
	}

	switch parts[0] {
	case "MC":
		return parseMC(parts[1])
	case "FREE":
		if len(parts) != 3 {
			return nil, fmt.Errorf("malformed free: %s", line)
		}
		return parseFree(parts[1], parts[2])
	default:
		return nil, fmt.Errorf("unknown answer type: %s", parts[0])
	}
}

type Section struct {
	answers []Answer
}

func newSection() Section {
	return Section{[]Answer{}}
}

func (s Section) normalise(path string) []string {
	contents, err := files.ReadFile(path)
	if err != nil {
		return []string{}
	}
	separators := []string{
		",", ";", ".", ":", "/",
		"\\", "|", "||"}
	// try different separators
	for _, sep := range separators {
		parts := strings.Split(contents, sep)
		if len(parts) > 1 {
			return parts
		}
	}
	parts := strings.Split(contents, "\n")
	// \n at end of file!
	return parts[:len(parts)-1]
}

func (s Section) Grade(path string, player string, section int) int {
	subs := s.normalise(path)
	lsubs := len(subs)
	la := len(s.answers)
	if lsubs != la {
		fmt.Printf("FLAG! %s in section %d has %d submissions but section requires %d\n", player, section, lsubs, la)
	}

	score := 0
	for i := 0; i < la && i < lsubs; i++ {
		score += s.answers[i].Grade(subs[i], player, section, i+1)
	}
	return score
}

func parseSection(lines []string, from int) (Section, int, error) {
	res := newSection()
	i := from
	for i < len(lines) {
		line := lines[i]
		if line == "" {
			break
		}
		ans, err := parseAnswer(line)
		if err != nil {
			return res, i, err
		}
		res.answers = append(res.answers, ans)
		i++
	}
	return res, i, nil
}

type Grader struct {
	sections []Section
}

func (g Grader) Grade(subPath string, section int) error {
	if section > len(g.sections) || section < 0 {
		return fmt.Errorf("invalid section number: %d", section)
	}
	secname := fmt.Sprintf("section%d", section)
	entries, err := os.ReadDir(filepath.Join(subPath, secname))
	if err != nil {
		return err
	}

	scores := map[string]int{}
	for _, e := range entries {
		player := e.Name()
		if player == "__score" {
			continue
		}
		secPath := filepath.Join(subPath, secname, player)
		scores[player] = g.sections[section-1].Grade(secPath, player, section)
	}

	scorePath := filepath.Join(subPath, secname, "__score")
	if files.Exists(scorePath) {
		err := os.Remove(scorePath)
		if err != nil {
			return err
		}
	}

	f, err := os.Create(scorePath)
	if err != nil {
		return err
	}

	defer f.Close()

	for player, score := range scores {
		_, err := f.WriteString(fmt.Sprintf("%s:%d\n", player, score))
		if err != nil {
			return err
		}
	}
	return nil
}

type pScore struct {
	player string
	score  int
}

func (g Grader) PrintScores(subPath string) error {
	secEntries, err := os.ReadDir(subPath)
	if err != nil {
		return err
	}

	scores := map[string]int{}
	for _, e := range secEntries {
		section := e.Name()

		secPath := filepath.Join(subPath, section, "__score")
		if !files.Exists(secPath) {
			continue
		}
		lines, err := files.ReadFileLines(secPath)
		if err != nil {
			return fmt.Errorf("error in %s: %v", section, err)
		}

		for _, line := range lines {
			parts := strings.Split(line, ":")
			if len(parts) != 2 {
				return fmt.Errorf("invalid score line in %s: %s", section, line)
			}
			p := parts[0]
			sc, err := strconv.Atoi(parts[1])
			if err != nil {
				return fmt.Errorf("error in %s: %v", section, err)
			}
			if _, exists := scores[p]; !exists {
				scores[p] = 0
			}
			scores[p] += sc
		}

	}
	pScores := []pScore{}
	for p, score := range scores {
		pScores = append(pScores, pScore{p, score})
	}
	sort.Slice(pScores, func(i, j int) bool {
		return pScores[i].score > pScores[j].score
	})

	i := 0
	for i < len(pScores) {
		rank := i + 1
		outStr := fmt.Sprintf("%d. ", rank)
		cScore := pScores[i].score
		rankPs := []string{}
		for i < len(pScores) && pScores[i].score == cScore {
			rankPs = append(rankPs, pScores[i].player)
			i++
		}

		// go back one
		i--
		outStr += strings.Join(rankPs, ", ")
		outStr += fmt.Sprintf(" (%d)", cScore)
		fmt.Println(outStr)
		i++
	}
	return nil
}

func newGrader() Grader {
	return Grader{[]Section{}}
}

func parseAnswers(lines []string) (Grader, error) {
	res := newGrader()
	i := 0
	for i < len(lines) {
		section, ni, err := parseSection(lines, i)
		if err != nil {
			return res, err
		}
		res.sections = append(res.sections, section)
		i = ni + 1
	}
	return res, nil
}

func FromFile(path string) (Grader, error) {
	answersStr, err := files.ReadFileLines(path)
	if err != nil {
		var grader Grader
		return grader, fmt.Errorf("could not read file: %v", err)
	}
	return parseAnswers(answersStr)
}
