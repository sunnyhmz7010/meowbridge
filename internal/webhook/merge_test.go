package webhook

import "testing"

func TestMergeFieldPrecedence(t *testing.T) {
	final, err := Merge(
		ParsedMessage{
			Title:   "parsed title",
			Msg:     "parsed msg",
			URL:     "https://parsed.test",
			ImgURL:  "https://parsed.test/icon.png",
			MsgType: "markdown",
		},
		EndpointDefaults{
			DefaultTitle:  "default title",
			MsgType:       "text",
			HTMLHeight:    200,
			DefaultURL:    "https://default.test",
			DefaultImgURL: "https://default.test/icon.png",
		},
		QueryOverrides{
			Title:      "query title",
			MsgType:    "html",
			HTMLHeight: 500,
		},
	)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if final.Title != "query title" || final.Msg != "parsed msg" || final.MsgType != "html" || final.HTMLHeight != 500 {
		t.Fatalf("final = %#v", final)
	}
}

func TestMergeRejectsEmptyMessage(t *testing.T) {
	_, err := Merge(ParsedMessage{}, EndpointDefaults{}, QueryOverrides{})
	if err == nil {
		t.Fatal("expected empty message error")
	}
}
