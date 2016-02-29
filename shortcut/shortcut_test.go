package shortcut

import (
	"flag"
	"os"
	"reflect"
	"testing"
)

var BLEVE_TEST_FILE = "testfile.bleve"

var testIndex Index

func assertEqual(t *testing.T, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v was %v", expected, actual)
	}
}

func TestRetrieveShortcutSimpleRed(t *testing.T) {
	res, sole, err := testIndex.FindShortcut("red")
	assertEqual(t, true, sole)
	assertEqual(t, nil, err)
	assertEqual(t, 1, len(res))
	assertEqual(t, "http://reddit.com", res[0].Url)
}

func TestRetrieveShortcutSimpleBlue(t *testing.T) {
	res, sole, err := testIndex.FindShortcut("blue")
	assertEqual(t, true, sole)
	assertEqual(t, nil, err)
	assertEqual(t, 1, len(res))
	assertEqual(t, "http://bluemoon.org", res[0].Url)
}

func TestRetrieveShortcutSimpleBlueFull(t *testing.T) {
	res, sole, err := testIndex.FindShortcut("blue-is-more-violet-than-red")
	assertEqual(t, true, sole)
	assertEqual(t, nil, err)
	assertEqual(t, 1, len(res))
	assertEqual(t, "http://bluemoon.org", res[0].Url)
}

func TestGetShortcutDescription(t *testing.T) {
	res, sole, err := testIndex.FindShortcut("page of")
	assertEqual(t, false, sole)
	assertEqual(t, nil, err)
	assertEqual(t, 1, len(res))
	assertEqual(t, "http://reddit.com", res[0].Url)
}

func TestGetShortcutDescriptionDisjoint(t *testing.T) {
	res, sole, err := testIndex.FindShortcut("internet page")
	assertEqual(t, false, sole)
	assertEqual(t, nil, err)
	assertEqual(t, 1, len(res))
	assertEqual(t, "http://reddit.com", res[0].Url)
}

func TestGetShortcutsDescription(t *testing.T) {
	res, sole, err := testIndex.FindShortcut("front")
	assertEqual(t, false, sole)
	assertEqual(t, nil, err)
	assertEqual(t, 2, len(res))
	assertEqual(t, "http://reddit.com", res[0].Url)
}

func TestGetDuplicateByDescription(t *testing.T) {
	res, sole, err := testIndex.FindShortcut("duplicate")
	assertEqual(t, false, sole)
	assertEqual(t, nil, err)
	assertEqual(t, 2, len(res))
	//	assertEqual(t, "http://reddit.com", res[0].Url)
}

func TestMain(m *testing.M) {
	flag.Parse()
	if os.RemoveAll(BLEVE_TEST_FILE) != nil {
		panic("Something went wrong with the file system.")
	}

	testIndex = NewIndex(BLEVE_TEST_FILE)

	testIndex.AddShortcut(Shortcut{
		ShortForm:   "blue",
		Url:         "bluemoon.org",
		Description: "A total dream of a website--Blue Mood\n\n James Front.",
	})

	testIndex.AddShortcut(Shortcut{
		ShortForm:   "red",
		Url:         "reddit.com",
		Description: "Reddit.com--The front page of the internet.",
	})

	testIndex.AddShortcut(Shortcut{
		ShortForm:   "blue-is-more-violet-than-red",
		Url:         "bluemoon.org",
		Description: "This is a duplicate",
	})

	testIndex.AddShortcut(Shortcut{
		ShortForm:   "2016-payroll",
		Url:         "ftp://myserv.org/files/secretfile",
		Description: "This is a duplicate",
	})

	os.Exit(m.Run())
}
