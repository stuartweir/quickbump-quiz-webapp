package main

import (
    //"./quickbumplib"
    "bytes"
    "encoding/json"
    "io/ioutil"
    "net/http"
    "net/http/httptest"
    "reflect"
    "testing"
    "time"
    "strconv"
)

var database = NewMemDb()
var handler, _ = NewQuickBumpHandler(database, "")
var server = httptest.NewServer(handler)

// eh, used for generating question names ...
var namecounter = func() (func() string) {
    idx := 0
    return func() string {
        idx += 1
        return strconv.Itoa(idx)
    }
}()

var inAnHour = time.Now().Add(time.Hour)
var inAnHourStr = func() string {
    data, err := json.Marshal(inAnHour)
    if err != nil {
        panic(err)
    }
    return string(data)
}()

//// General Test Functions //////////////////////////////////////////////////

func readBody(resp *http.Response) (string, error) {
    data, err := ioutil.ReadAll(resp.Body)
    return string(data), err
}

func GetThing(t *testing.T, url string, thing interface{}) {
    resp, err := http.Get(server.URL + url)
    if err != nil {
        t.Fatal(err)
    }
    data, _ := ioutil.ReadAll(resp.Body)
    if resp.StatusCode != 200 {
        t.Fatalf("Unexpected status code %d. Body: %s", resp.StatusCode, data)
    }

    if err = json.Unmarshal(data, thing); err != nil {
        t.Log("Unmarshalling: ", string(data))
        t.Fatal(err)
    }
}

func PostData(url string, data []byte) (*http.Response, error) {
    return http.Post(
        server.URL + url,
        "application/json",
        bytes.NewBuffer(data),
    )
}

func PostThing(t *testing.T, url string, thing interface{}) (*http.Response, error) {
    data, err := json.Marshal(thing)
    if err != nil {
        t.Log("Marshalling: ", thing)
        t.Fatal(err)
    }
    return PostData(url, data)
}

//// Question Test Functions ////////////////////////////////////////////////

func GetQuestion(t *testing.T, id Id) *Question {
    q := &Question{}
    GetThing(t, "/question/"+string(id), q)
    return q
}

//// Answer Test Functions ///////////////////////////////////////////////////

func GetAnswer(t *testing.T, id Id) *Answer {
    a := &Answer{}
    GetThing(t, "/answer/"+string(id), a)
    return a
}

func GetAnswersForQuestion(t *testing.T, question_id Id) []*Answer {
    things := make(map[string]*Answer)
    GetThing(t, "/answer?Question="+string(question_id), &things)
    list := make([]*Answer, len(things))
    idx := 0
    for _, answer := range things {
        list[idx] = answer
    }
    return list
}

//// Rawr ////////////////////////////////////////////////////////////////////

func TestQuestion(t *testing.T) {
    input_info := TextInfo{"Do you know where your towel is?", 99, 0}
    q := &QuestionCreationData{time.Unix(0, 0), namecounter(), PollData{TextPoll, input_info}}
    resp, _ := PostThing(t, "/question", q)
    body, _ := readBody(resp)
    id := Id(body)
    if id == NullId {
        t.Fatal("Setup for TestQuestion goofed up")
    }

    question := GetQuestion(t, id)
    if question.Data.Mode != TextPoll {
        t.Fatal("%s != %s", TextPoll, question.Data.Mode)
    }
    if info, ok := question.Data.Info.(*TextInfo); !ok {
        t.Fatal("nope! ", question.Data.Info)
    } else if !reflect.DeepEqual(info, &input_info) {
        t.Fatalf("%s != %s", info, &input_info)
    }
}

func TestAnswerPostGet(t *testing.T) {
    data := PollData{TextPoll, &TextInfo{"Hey, you sass that hoopy Ford Prefect?", 0, 4}}
    question_id, err := database.Insert(NewQuestion(time.Unix(0, 0), namecounter(), data))
    if err != nil {
        t.Fatal("Problem setting up TestAnswer", err)
    }

    resp, _ := PostData("/answer", []byte(`{
        "QuestionId": "` + string(question_id) + `",
        "Response": "42"
    }`))
    body, _ := readBody(resp)
    id := Id(body)
    if id == NullId {
        t.Fatal("Setup for TestAnswer goofed up")
    }

    answer := GetAnswer(t, id)
    if answer.Response != "42" {
        t.Fatal("nope! ", answer)
    }

    answers := GetAnswersForQuestion(t, question_id)
    if len(answers) != 1 || answers[0].Response != "42" {
        t.Fatal("nope! ", answers)
    }

    answers = GetAnswersForQuestion(t, "lksdfjksldjfkl")
    if len(answers) != 0 {
        t.Fatal("nope nope nope! ", answers)
    }
}

// Validity testing
// ... this is actually testing the handler as well as the model logic;
// they should be tested independently ... :|

var QuestionValidationData = map[string]bool{
    `{
        "End": ` + inAnHourStr + `,
        "Data": {
            "Mode": "NOPE",
            "Info": {}
        }
    }`: false,
    `{
        "End": ` + inAnHourStr + `,
        "Data": {
            "Mode": "TEXT",
            "Info": {
                "Question": {"Derp": 42},
                "WordLimit": 0,
                "CharacterLimit": 1
            }
        }
    }`: false,
    `{
        "End": ` + inAnHourStr + `,
        "Data": {
            "Mode": "TEXT",
            "Info": {
                "Question": "How many chucks could a wood chuck chuck if a wood chuck could chuck wood?",
                "WordLimit": 0,
                "CharacterLimit": 1
            }
        }
    }`: true,
    `{
        "End": ` + inAnHourStr + `,
        "Data": {
            "Mode": "TEXT",
            "Info": {
                "Question": "",
                "WordLimit": 0,
                "CharacterLimit": 1
            }
        }
    }`: false,
    `{
        "End": ` + inAnHourStr + `,
        "Data": {
            "Mode": "CHOICE",
            "Info": {
                "Question": "?",
                "MinChoices": 0,
                "MaxChoices": 1,
                "Choices": [42]
            }
        }
    }`: false,
}

func TestQuestionValidation(t *testing.T) {
    for data, pass := range QuestionValidationData {
        resp, _ := PostData("/question", []byte(data))
        if resp.StatusCode == 200 != pass {
            body, _ := readBody(resp)
            t.Errorf(
                "Validation failure!\nExpected: %t\nData: %s\nBody: %s",
                pass,
                data,
                body,
            )
        }
    }
}

var AnswerValidationData = map[*Question]map[string]bool{
    NewQuestion(
        inAnHour,
        namecounter(),
        PollData{ChoicePoll, ChoiceInfo{"?", 1, 2, []string{"A", "B", "C"}}},
    ): map[string]bool{
        `{"Yarrr": "Don't be ridiculous"}`: false,
        `[]`: false,
        `[0, 1, 2]`: false,
        `[0, 0]`: false,
        `[2, 1]`: true,
        `[0]`: true,
        `[3]`: false,
    },

    NewQuestion(
        inAnHour,
        namecounter(),
        PollData{TextPoll, TextInfo{"?", 0, 4}},
    ): map[string]bool{
        "[]": false,
        `"3!?"`: true,
        `"four"`: true,
        `"five!"`: false,
    },

    NewQuestion(
        inAnHour,
        namecounter(),
        PollData{TextPoll, TextInfo{"?", 2, 0}},
    ): map[string]bool{
        `"one"`: true,
        `"  one   two  "`: true,
        `"one.two?three"`: false,
        `"one-two1three"`: false,
    },
}

func TestAnswerValidation(t *testing.T) {
    for question, stuff := range AnswerValidationData {
        question_id, err := database.Insert(question)
        if err != nil {
            t.Fatal("Problem setting up TestAnswerValidation", err)
        }

        for response, result := range stuff {
            resp, _ := PostData("/answer", []byte(
                `{
                    "QuestionId": "` + string(question_id) + `",
                    "Response": ` + response + `
                }`))
            if resp.StatusCode == 200 != result {
                body, _ := readBody(resp)
                t.Errorf(
                    "Validation failure!\nExpected: %t\nData: %s\nBody: %s",
                    result,
                    response,
                    body,
                )
            }
        }
    }
}
