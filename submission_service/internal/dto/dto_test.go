package dto

import "testing"

func TestSubmitRequest_Validate(t *testing.T) {
	tests := []struct {
		name         string
		req          SubmitRequest
		wantErrField string
		wantErrMsg   string
	}{
		{
			name:         "valid",
			req:          SubmitRequest{ProblemID: 1, LanguageID: 1, Code: "console.log(1)"},
			wantErrField: "",
			wantErrMsg:   "",
		},
		{
			name:         "missing problemId",
			req:          SubmitRequest{LanguageID: 1, Code: "console.log(1)"},
			wantErrField: "problemId",
			wantErrMsg:   "Problem ID is required",
		},
		{
			name:         "missing languageId",
			req:          SubmitRequest{ProblemID: 1, Code: "console.log(1)"},
			wantErrField: "languageId",
			wantErrMsg:   "Language ID is required",
		},
		{
			name:         "empty code",
			req:          SubmitRequest{ProblemID: 1, LanguageID: 1, Code: ""},
			wantErrField: "code",
			wantErrMsg:   "Code must not be empty",
		},
		{
			name:         "whitespace-only code",
			req:          SubmitRequest{ProblemID: 1, LanguageID: 1, Code: "   \n\t  "},
			wantErrField: "code",
			wantErrMsg:   "Code must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, msg := tt.req.Validate()
			if field != tt.wantErrField {
				t.Errorf("Validate() field = %v, want %v", field, tt.wantErrField)
			}
			if msg != tt.wantErrMsg {
				t.Errorf("Validate() msg = %v, want %v", msg, tt.wantErrMsg)
			}
		})
	}
}
