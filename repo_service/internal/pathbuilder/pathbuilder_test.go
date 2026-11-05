package pathbuilder

import (
	"testing"
)

func TestBuildRepoName(t *testing.T) {
	got := BuildRepoName("javascript")
	want := "javascript-RinnoGen"
	if got != want {
		t.Errorf("BuildRepoName = %q, want %q", got, want)
	}
}

func TestBuildFilePath(t *testing.T) {
	ctx := CurriculumContext{
		SubjectSlug:  "python",
		SessionOrder: 1,
		LessonOrder:  2,
		ProblemOrder: 3,
		ProblemSlug:  "two-sum",
	}
	got := BuildFilePath(ctx, "solution.py")
	want := "python/Session-01/Lesson-02/Problem-03-two-sum/solution.py"
	if got != want {
		t.Errorf("BuildFilePath = %q, want %q", got, want)
	}
}

func TestBuildFilePath_ZeroPadding(t *testing.T) {
	ctx := CurriculumContext{
		SubjectSlug:  "cpp",
		SessionOrder: 10,
		LessonOrder:  5,
		ProblemOrder: 99,
		ProblemSlug:  "fibonacci",
	}
	got := BuildFilePath(ctx, "main.cpp")
	want := "cpp/Session-10/Lesson-05/Problem-99-fibonacci/main.cpp"
	if got != want {
		t.Errorf("BuildFilePath = %q, want %q", got, want)
	}
}

func TestGenerateCommitSHA(t *testing.T) {
	sha := GenerateCommitSHA()
	if len(sha) != 40 {
		t.Errorf("GenerateCommitSHA length = %d, want 40", len(sha))
	}
	for _, c := range sha {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("GenerateCommitSHA contains invalid hex char: %c", c)
			break
		}
	}
}
