package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "math/rand"
    "net/http"
    "net/url"
    "reflect"
    "strings"
    "time"
)

var Letters = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}

type QuestionCreationData struct {
    End  time.Time
    Name    string
    Data PollData
}

type AnswerCreationData struct {
    // TODO, rename this to Question and update docs
    QuestionId  Id
    Response    json.RawMessage
}

// Object responsible for handling http requests
// Implements http.Handler
type QuickBumpHandler struct {
    Db          Database
    Wordlist    []string
}

// Given url.Values and an interface, this will do its best to resolve values
// on the query string based on the structure of the given model
// todo, this might be easier by json encoding the query and putting it in
// the request body
func rebuildQuery(query url.Values, model Thing) (QuerySpec, error) {
    modeltype := reflect.TypeOf(model)
    if modeltype.Kind() == reflect.Ptr {
        modeltype = modeltype.Elem()
    }
    spec := make(QuerySpec)
    for name, valuelist := range query {
        field, found := modeltype.FieldByName(name)
        if !found {
            return nil, fmt.Errorf("Field `%s' not found on `%s'", name, modeltype.Name())
        }
        if kind := field.Type.Kind(); kind != reflect.String {
            return nil, fmt.Errorf("Query on unsupported field kind `%s'", kind)
        }
        spec[name] = make([]interface{}, len(valuelist))
        for idx, value := range valuelist {
            holder := reflect.New(field.Type).Elem()
            holder.SetString(value)
            spec[name][idx] = holder.Interface()
        }
    }
    return spec, nil
}

func NewQuickBumpHandler(db Database, wordfile string) (*QuickBumpHandler, error) {
    var wordlist []string
    if wordfile != "" {
        data, err := ioutil.ReadFile(wordfile)
        if err != nil {
            return nil, err
        }
        wordlist = strings.Split(string(data), "\n")
    }
    return &QuickBumpHandler{db, wordlist}, nil
}

func sample(count int, source []string) []string {
    picks := make([]string, count)
    for x := 0; x < count; x++ {
        pick := source[rand.Int() % len(source)]
        picks = append(picks, pick)
    }
    return picks
}

func (h *QuickBumpHandler) generateQuestionName() string {
    if len(h.Wordlist) != 0 {
        return strings.Join(sample(3, h.Wordlist), " ")
    }
    return strings.Join(sample(6, Letters), "")
}

func (h *QuickBumpHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    url := req.URL.Path

    if url == "" || url == "/" {
        h.ServeDocumentation(w, req)
        return
    }

    chunks := strings.Split(url[1:], "/")

    if !((len(chunks) == 1 && (req.Method == "POST" || req.Method == "GET")) || len(chunks) == 2) {
        http.NotFound(w, req)
        return
    }

    var err error
    switch req.Method {
    case "GET":
        // URLs in the form /<model_type> /<model_type>/<id>
        err = h.ServeGet(w, req, chunks)
    case "POST":
        // URLs in the form /<model_name>
        err = h.ServePost(w, req, chunks)
    //case "PUT":
    //    // URLs in the form /<model_type>/<id>
    //    err = h.ServePut(w, req, chunks)
    //case "DELETE":
    //    // URLs in the form /<model_type>/<id>
    //    err = h.ServeDelete(w, req, chunks)
    default:
        http.Error(w,
            "Only GET and POST requests are handled by this API",
            http.StatusMethodNotAllowed)
    }
    if err != nil {
        http.Error(w, err.Error(), 500)
    }
}

func (h *QuickBumpHandler) ServeDocumentation(w http.ResponseWriter, req *http.Request) {
    if req.Method != "GET" {
        http.Error(w, "Only GET requests are accepted at this URL",
            http.StatusMethodNotAllowed)
        return
    }
    w.Write([]byte("todo, add documentation!"))
}

func (h *QuickBumpHandler) ServeGet(w http.ResponseWriter, req *http.Request, chunks []string) error {
    var err error
    var model Thing

    switch chunks[0] {
    case "question":
        model = &Question{}
    case "answer":
        model = &Answer{}
    default:
        return fmt.Errorf("Unknown data thingy `%s'", chunks[0])
    }

    switch len(chunks) {
    case 1:
        queryvalues := req.URL.Query()
        if len(queryvalues) == 0 {
            return fmt.Errorf("Querying without filters is not permitted")
        }
        queryspec, err := rebuildQuery(req.URL.Query(), model)
        if err != nil {
            return err
        }
        results := make(map[Id]Thing)
        each := h.Db.Query(queryspec , model)
        for {
            var model Thing // todo, this is shady
            if chunks[0] == "question" {
                model = &Question{}
            } else {
                model = &Answer{}
            }
            id, err := each(model)
            if err != nil {
                return err
            }
            if id == "" {
                break
            }
            results[id] = model
        }
        model = results // sure ......

    case 2:
        id := Id(chunks[1])
        err = h.Db.Load(id, model)
        if err == RecordNotFound {
            http.Error(w, err.Error(), 404)
            return nil
        } else if err != nil {
            return err
        }

    default:
        return fmt.Errorf("You aren't supposed to reach this code path, go find a programmer")
    }

    data, err := json.Marshal(model)
    if err != nil {
        return err
    }
    w.Write(data)
    return nil
}

func (h *QuickBumpHandler) ServePost(w http.ResponseWriter, req *http.Request, chunks []string) error {
    data, err := ioutil.ReadAll(req.Body)
    if err != nil {
        return err
    }
    log.Print("POSTed ", string(data))

    var object Thing

    switch chunks[0] {
    case "question":
        xq := QuestionCreationData{}
        if err := json.Unmarshal(data, &xq); err != nil {
            return err
        }

        if xq.Name == "" {
            xq.Name = h.generateQuestionName()
        } else {
            // Yeah, I know, this logic should be somewhere else entirely ...
            // what are you going to do about it?
            // Man this query stuff is ugly ...
            holder := Question{}
            r := h.Db.Query(QuerySpec{"Name": []interface{}{xq.Name}}, holder)
            if id, err := r(holder); err != nil {
                log.Print("An error occured while trying to check for dupes of the question name `%s': %s", xq.Name, err)
                return err
            } else if id != NullId {
                // todo Write a test for this
                return fmt.Errorf("Validation error: Name `%s' already in use", xq.Name)
            }
        }

        q := NewQuestion(xq.End, xq.Name, xq.Data)
        if err := q.Validate(); err != nil {
            return fmt.Errorf("Validation error: %s", err)
        }
        object = q

    case "answer":
        xa := AnswerCreationData{}
        if err := json.Unmarshal(data, &xa); err != nil {
            return err
        }
        q := &Question{}
        if err := h.Db.Load(xa.QuestionId, q); err != nil {
            return err
        }
        var resp PollResponse
        switch q.Data.Mode {
        case ChoicePoll:
            hew := new([]uint)
            if err := json.Unmarshal(xa.Response, hew); err != nil {
                return err
            }
            resp = *hew
        case TextPoll:
            hew := new(string)
            if err := json.Unmarshal(xa.Response, hew); err != nil {
                return err
            }
            resp = *hew
        default:
            return UnknownPollMode(q.Data.Mode)
        }
        a := NewAnswer(xa.QuestionId, resp)
        if err := a.Validate(q); err != nil {
            return err
        }
        object = a

    default:
        http.NotFound(w, req)
        return nil
    }

    id, err := h.Db.Insert(object)
    if err != nil {
        return err
    }

    log.Print("Inserted ", id)
    w.Write([]byte(id))
    return nil
}

//func (h *QuickBumpHandler) ServePut(w http.ResponseWriter, req *http.Request, chunks []string) error {
//    w.Write([]byte("todo"))
//    return nil
//}
//
//func (h *QuickBumpHandler) ServeDelete(w http.ResponseWriter, req *http.Request, chunks []string) error {
//    w.Write([]byte("todo"))
//    return nil
//}
