// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package vdir

import "github.com/sarifsystems/sarif/pkg/schema"

type Card struct {
	schema.Thing

	Profile       string          `vdir:"vcard,profile" json:"-"`
	FormattedName string          `vdir:"fn" json:"fn,omitempty"`
	Name          Name            `vdir:"n" json:"hasName,omitempty"`
	NickName      []string        `json:"nickname,omitempty"`
	Birthday      string          `vdir:"bday" json:"bday,omitempty"`
	Addresses     []Address       `vdir:"adr" json:"hasAddress,omitempty"`
	Telephones    []TypedValue    `vdir:"tel" json:"hasTelephone,omitempty"`
	Email         []TypedValue    `json:"hasEmail,omitempty"`
	Url           []TypedResource `json:"hasUrl,omitempty"`
	Title         string          `json:"title,omitempty"`
	Role          string          `json:"role,omitempty"`
	Org           string          `json:"org,omitempty"`
	Categories    []string        `json:"category,omitempty"`
	Note          string          `json:"note,omitempty"`

	Rev    string `json:"-"`
	ProdId string `json:"-"`
	Uid    string `json:"uid,omitempty"`

	IMPP []TypedResource `json:"hasInstantMessaging,omitempty"`
}

func (c Card) String() string {
	return "vCard of " + c.FormattedName
}

type Name struct {
	FamilyName        []string `json:"family-name,omitempty"`
	GivenName         []string `json:"given-name,omitempty"`
	AdditionalNames   []string `json:"additional-name,omitempty"`
	HonorificNames    []string `json:"honorific-prefix,omitempty"`
	HonorificSuffixes []string `json:"honorific-suffix,omitempty"`
}

type Address struct {
	Type            []string `vdir:",param" json:"@type,omitempty"`
	Label           string   `vdir:",param" json:"label,omitempty"`
	PostOfficeBox   string   `json:"-"`
	ExtendedAddress string   `json:"-"`
	Street          string   `json:"street,omitempty"`
	Locality        string   `json:"locality,omitempty"`
	Region          string   `json:"region,omitempty"`
	PostalCode      string   `json:"postal-code,omitempty"`
	CountryName     string   `json:"country-name,omitempty"`
}

type TypedValue struct {
	Type  []string `vdir:",param" json:"@type,omitempty"`
	Value string   `json:"hasValue"`
}

type TypedResource struct {
	Type  []string `vdir:",param" json:"@type,omitempty"`
	Value string   `json:"@id"`
}
