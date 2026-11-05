package pathbuilder

import (
	"crypto/rand"
	"fmt"
)

// CurriculumContext holds the curriculum hierarchy info needed to build a file path.
type CurriculumContext struct {
	SubjectSlug  string
	SubjectID    int
	SessionOrder int
	LessonOrder  int
	ProblemOrder int
	ProblemSlug  string
}

// BuildRepoName returns the repository name for a given subject.
// Format: <subjectSlug>-RinnoGen
func BuildRepoName(subjectSlug string) string {
	return subjectSlug + "-RinnoGen"
}

// BuildFilePath returns the full file path for a submission within a repository.
// Format: <subjectSlug>/Session-<NN>/Lesson-<NN>/Problem-<NN>-<problemSlug>/<fileName>
func BuildFilePath(ctx CurriculumContext, fileName string) string {
	return fmt.Sprintf("%s/Session-%02d/Lesson-%02d/Problem-%02d-%s/%s",
		ctx.SubjectSlug,
		ctx.SessionOrder,
		ctx.LessonOrder,
		ctx.ProblemOrder,
		ctx.ProblemSlug,
		fileName,
	)
}

// GenerateCommitSHA returns a random 40-character hex string to use as a mock commit SHA.
func GenerateCommitSHA() string {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		// Fallback to zeroes in the extremely unlikely case crypto/rand fails
		return "0000000000000000000000000000000000000000"
	}
	return fmt.Sprintf("%x", b)
}
