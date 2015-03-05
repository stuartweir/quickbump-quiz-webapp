package main

import (
    "time"
    "encoding/json"
    "fmt"
)

type PollMode string

var (
    ChoicePoll  PollMode = "CHOICE"
    TextPoll    PollMode = "TEXT"
)

var (
    EmptyQuestionErr = fmt.Errorf("Blank question text")
)

type UnknownPollMode PollMode

func (mode UnknownPollMode) Error() string {
    return fmt.Sprintf("Unknown poll mode `%s'", string(mode))
}

type ChoiceInfo struct {
    Question    string
    MinChoices  uint
    MaxChoices  uint
    Choices     []string
}

type TextInfo struct {
    Question   string
    WordLimit  uint
    CharLimit  uint
}

type PollData struct {
    Mode PollMode
    Info interface{}
}

func (poll *PollData) UnmarshalJSON(data []byte) error {
    holder := struct{
        Mode PollMode
        Info json.RawMessage
    }{}

    if err := json.Unmarshal(data, &holder); err != nil {
        return err
    }
    poll.Mode = holder.Mode

    var info interface{}
    switch poll.Mode {
    case ChoicePoll:
        info = &ChoiceInfo{}
    case TextPoll:
        info = &TextInfo{}
    default:
        return UnknownPollMode(poll.Mode)
    }
    poll.Info = info
    return json.Unmarshal(holder.Info, info)
}

type Question struct {
    Created time.Time
    End     time.Time
    Name    string
    Data    PollData
}

func (q *Question) Validate() error {
    if q.Name == "" { // todo, write a test for this
        return fmt.Errorf("Question Name must be nonempty")
    }
    switch q.Data.Mode {
    case ChoicePoll:
        info := q.Data.Info.(*ChoiceInfo)
        if info.Question == "" {
            return EmptyQuestionErr
        }
    case TextPoll:
        info := q.Data.Info.(*TextInfo)
        if info.Question == "" {
            return EmptyQuestionErr
        }
    default:
        return UnknownPollMode(q.Data.Mode)
    }
    return nil
}

func NewQuestion(end time.Time, name string, data PollData) *Question {
    return &Question{time.Now(), end, name, data}
}
