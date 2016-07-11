// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package schema

import "time"

type Action struct {
	*Thing
	Reply         string      `json:"reply,omitempty"`
	ReplyNegative string      `json:"reply_neg,omitempty"`
	Name          string      `json:"name,omitempty"`
	Payload       interface{} `json:"p,omitempty"`
	Style         string      `json:"style,omitempty"`
}

type MultipleChoiceAction struct {
	Action
	Choices map[string]interface{} `json:"choices,omitempty"`
}

type TextEntryAction Action
type ConfirmAction Action
type DeleteAction Action
type CancelAction Action

type Attachment struct {
	*Thing

	Fallback string `json:"fallback,omitempty"`
	Color    string `json:"color,omitempty"`
	Pretext  string `json:"pretext,omitempty"`

	AuthorName string `json:"author_name,omitempty"`
	AuthorLink string `json:"author_link,omitempty"`
	AuthorIcon string `json:"author_icon,omitempty"`

	Title     string            `json:"title,omitempty"`
	TitleLink string            `json:"title_link,omitempty"`
	Fields    []AttachmentField `json:"fields,omitempty"`
	Text      string            `json:"text,omitempty"`

	ImageUrl   string `json:"image_url,omitempty"`
	ThumbUrl   string `json:"thumb_url,omitempty"`
	Footer     string `json:"footer,omitempty"`
	FooterIcon string `json:"footer_icon,omitempty"`

	Time    time.Time     `json:"time,omitempty"`
	Actions []interface{} `json:"actions,omitempty"`
}

type AttachmentField struct {
	Title string `json:"title,omitempty"`
	Value string `json:"value,omitempty"`
	Short bool   `json:"short,omitempty"`
}
