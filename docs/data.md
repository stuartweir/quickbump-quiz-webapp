*All this is subject to change*

# Data Models

These are descriptions of the layout of objects which are put into databases used by QuickBump. Exaclty how they appear in a particular database is up to our implementation that adapts QuickBump for that database.

    ModelId := string

Ids are not placed on models by QuickBump. The database implementation is responsible for generating keys and associating them with its records however it pleases.

The `Created` attributes of each model do not need to be provided in the POST request.

## Questions

The value of the `Data` attribute describes the poll mode and the details such as the poll question.

    Question := {
        Created:    Time,
        End:        Time,
        Data:       PollData,
    }

## Answers

The type of the `Response` attribute is dependent on the poll mode for the associated `Question`.

    Answer := {
        QuestionId: ModelId,
        Created:    Time,
        Response:   PollResponse,
    }

## Poll Modes and Mode Data

The value for `PollData` on `Question` determines the poll mode and its details. Below are definitions of valid structures for `PollData`. The structure of `Info`, as well as `PollResponse` on the `Answer` model, depends on the value of `Mode`.

### Multiple Choice

    PollData := {
        Mode:  "CHOICE",
        Info: {
            Question:   string,
            MinChoices: uint,
            MaxChoices: uint,
            Choices:    [string, ...],
        },
    }

    PollResponse := [uint, ...]

### Short-Answer

    PollData := {
        Mode: "TEXT",
        Info: {
            Question:       string,
            WordLimit:      uint,
            CharacterLimit: uint,
        },
    }

    PollResponse := string

## Example

    PrimeQuestion := Question{
        Created:    time.Date(2013, time.February, 14, 11, 23, 50, 0, time.UTC),
        End:        null,
        Data: {
            Mode: "CHOICE",
            Info: {
                Question:   "Which of the numbers listed are prime?",
                MinChoices: 0,
                MaxChoices: 4,
                Choices:    ["7", "9", "193", "199"],
            },
        },
    }

    CorrectAnswer := Question{
        QuestionId: "???",
        Created:    time.Date(2013, time.March, 03, 23, 42, 29, 0, time.UTC),
        Response:   [0, 2, 3],
    }
