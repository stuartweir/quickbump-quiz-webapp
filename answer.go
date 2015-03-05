package main

import (
    "fmt"
    "strings"
    "time"
    "unicode"
    "unicode/utf8"
)

type PollResponse interface{}

type Answer struct {
    Question    Id
    Created     time.Time
    Response    PollResponse
}

func wordcount(text string) (c uint) {
    c = uint(0)
    onword := false
    for _, r := range text {
        is := unicode.IsLetter(r)
        if !onword && onword != is {
            c++
        }
        onword = is
    }
    return
}

func isUnique(slice []uint) bool {
    for idx, elm := range slice {
        for _, other := range slice[idx+1:] {
            if elm == other {
                return false
            }
        }
    }
    return true
}

// I really hate the way I do this validation stuff ...
func (a *Answer) Validate(q *Question) error {
    if e := q.Validate(); e != nil {
        return e
    }

    switch q.Data.Mode {
    case ChoicePoll:
        qinfo := q.Data.Info.(*ChoiceInfo)
        resp := a.Response.([]uint)
        nresp := uint(len(resp))
        if nresp < qinfo.MinChoices {
            return fmt.Errorf("Too few choices")
        }
        if nresp  > qinfo.MaxChoices {
            return fmt.Errorf("Too many choices")
        }

        for _, v := range resp {
            if v >= uint(len(qinfo.Choices)) {
                return fmt.Errorf("Choice out of range")
            }
        }
        if !isUnique(resp) {
            return fmt.Errorf("Given choices must be unique")
        }

    case TextPoll:
        qinfo := q.Data.Info.(*TextInfo)
        resp := a.Response.(string)
        trimmed := strings.Trim(resp, " \n\r\t")
        if qinfo.WordLimit != 0 {
            count := wordcount(trimmed)
            if count > qinfo.WordLimit {
                return fmt.Errorf(
                    "Response exceeds word limit. Counted %d. Limit %d.", count, qinfo.WordLimit)
            }
        }
        if qinfo.CharLimit != 0 {
            count := uint(utf8.RuneCountInString(trimmed));
            if count > qinfo.CharLimit {
                return fmt.Errorf("Response exceeds character limit. Counted %d. Limit %d",
                    count, qinfo.CharLimit)
            }
        }
    }
    return nil
}

func NewAnswer(id Id, response PollResponse) *Answer {
    return &Answer{id, time.Now(), response}
}
