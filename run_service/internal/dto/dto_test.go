package dto

import "testing"

func TestRunRequest_Validate(t *testing.T) {
	tests := []struct {
		name         string
		req          RunRequest
		wantErrField string
		wantErrMsg   string
	}{
		{
			name:         "valid",
			req:          RunRequest{ProblemID: "00000000-0000-0000-0000-000000000000", LanguageID: "00000000-0000-0000-0000-000000000000", Code: "console.log(1)"},
			wantErrField: "",
			wantErrMsg:   "",
		},
		{
			name:         "missing problemId",
			req:          RunRequest{LanguageID: "00000000-0000-0000-0000-000000000000", Code: "console.log(1)"},
			wantErrField: "problemId",
			wantErrMsg:   "Invalid problem ID",
		},
		{
			name:         "missing languageId",
			req:          RunRequest{ProblemID: "00000000-0000-0000-0000-000000000000", Code: "console.log(1)"},
			wantErrField: "languageId",
			wantErrMsg:   "Invalid language ID",
		},
		{
			name:         "empty code",
			req:          RunRequest{ProblemID: "00000000-0000-0000-0000-000000000000", LanguageID: "00000000-0000-0000-0000-000000000000", Code: ""},
			wantErrField: "code",
			wantErrMsg:   "code must not be empty",
		},
		{
			name:         "whitespace-only code",
			req:          RunRequest{ProblemID: "00000000-0000-0000-0000-000000000000", LanguageID: "00000000-0000-0000-0000-000000000000", Code: "   \n\t  "},
			wantErrField: "code",
			wantErrMsg:   "code must not be empty",
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
