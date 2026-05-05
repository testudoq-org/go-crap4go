package bad_syntax

// This file contains a syntax error so that the pipeline can test
// graceful error handling.
func BrokenFunc( {
	return 42
}
