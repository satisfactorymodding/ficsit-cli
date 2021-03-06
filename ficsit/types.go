// Code generated by github.com/Khan/genqlient, DO NOT EDIT.

package ficsit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/satisfactorymodding/ficsit-cli/ficsit/utils"
)

// GetModMod includes the requested fields of the GraphQL type Mod.
type GetModMod struct {
	Id               string                    `json:"id"`
	Mod_reference    string                    `json:"mod_reference"`
	Name             string                    `json:"name"`
	Views            int                       `json:"views"`
	Downloads        int                       `json:"downloads"`
	Authors          []GetModModAuthorsUserMod `json:"authors"`
	Full_description string                    `json:"full_description"`
	Source_url       string                    `json:"source_url"`
	Created_at       time.Time                 `json:"-"`
}

// GetId returns GetModMod.Id, and is useful for accessing the field via an interface.
func (v *GetModMod) GetId() string { return v.Id }

// GetMod_reference returns GetModMod.Mod_reference, and is useful for accessing the field via an interface.
func (v *GetModMod) GetMod_reference() string { return v.Mod_reference }

// GetName returns GetModMod.Name, and is useful for accessing the field via an interface.
func (v *GetModMod) GetName() string { return v.Name }

// GetViews returns GetModMod.Views, and is useful for accessing the field via an interface.
func (v *GetModMod) GetViews() int { return v.Views }

// GetDownloads returns GetModMod.Downloads, and is useful for accessing the field via an interface.
func (v *GetModMod) GetDownloads() int { return v.Downloads }

// GetAuthors returns GetModMod.Authors, and is useful for accessing the field via an interface.
func (v *GetModMod) GetAuthors() []GetModModAuthorsUserMod { return v.Authors }

// GetFull_description returns GetModMod.Full_description, and is useful for accessing the field via an interface.
func (v *GetModMod) GetFull_description() string { return v.Full_description }

// GetSource_url returns GetModMod.Source_url, and is useful for accessing the field via an interface.
func (v *GetModMod) GetSource_url() string { return v.Source_url }

// GetCreated_at returns GetModMod.Created_at, and is useful for accessing the field via an interface.
func (v *GetModMod) GetCreated_at() time.Time { return v.Created_at }

func (v *GetModMod) UnmarshalJSON(b []byte) error {

	if string(b) == "null" {
		return nil
	}

	var firstPass struct {
		*GetModMod
		Created_at json.RawMessage `json:"created_at"`
		graphql.NoUnmarshalJSON
	}
	firstPass.GetModMod = v

	err := json.Unmarshal(b, &firstPass)
	if err != nil {
		return err
	}

	{
		dst := &v.Created_at
		src := firstPass.Created_at
		if len(src) != 0 && string(src) != "null" {
			err = utils.UnmarshalDateTime(
				src, dst)
			if err != nil {
				return fmt.Errorf(
					"Unable to unmarshal GetModMod.Created_at: %w", err)
			}
		}
	}
	return nil
}

type __premarshalGetModMod struct {
	Id string `json:"id"`

	Mod_reference string `json:"mod_reference"`

	Name string `json:"name"`

	Views int `json:"views"`

	Downloads int `json:"downloads"`

	Authors []GetModModAuthorsUserMod `json:"authors"`

	Full_description string `json:"full_description"`

	Source_url string `json:"source_url"`

	Created_at json.RawMessage `json:"created_at"`
}

func (v *GetModMod) MarshalJSON() ([]byte, error) {
	premarshaled, err := v.__premarshalJSON()
	if err != nil {
		return nil, err
	}
	return json.Marshal(premarshaled)
}

func (v *GetModMod) __premarshalJSON() (*__premarshalGetModMod, error) {
	var retval __premarshalGetModMod

	retval.Id = v.Id
	retval.Mod_reference = v.Mod_reference
	retval.Name = v.Name
	retval.Views = v.Views
	retval.Downloads = v.Downloads
	retval.Authors = v.Authors
	retval.Full_description = v.Full_description
	retval.Source_url = v.Source_url
	{

		dst := &retval.Created_at
		src := v.Created_at
		var err error
		*dst, err = json.Marshal(
			&src)
		if err != nil {
			return nil, fmt.Errorf(
				"Unable to marshal GetModMod.Created_at: %w", err)
		}
	}
	return &retval, nil
}

// GetModModAuthorsUserMod includes the requested fields of the GraphQL type UserMod.
type GetModModAuthorsUserMod struct {
	Role string                      `json:"role"`
	User GetModModAuthorsUserModUser `json:"user"`
}

// GetRole returns GetModModAuthorsUserMod.Role, and is useful for accessing the field via an interface.
func (v *GetModModAuthorsUserMod) GetRole() string { return v.Role }

// GetUser returns GetModModAuthorsUserMod.User, and is useful for accessing the field via an interface.
func (v *GetModModAuthorsUserMod) GetUser() GetModModAuthorsUserModUser { return v.User }

// GetModModAuthorsUserModUser includes the requested fields of the GraphQL type User.
type GetModModAuthorsUserModUser struct {
	Username string `json:"username"`
}

// GetUsername returns GetModModAuthorsUserModUser.Username, and is useful for accessing the field via an interface.
func (v *GetModModAuthorsUserModUser) GetUsername() string { return v.Username }

// GetModResponse is returned by GetMod on success.
type GetModResponse struct {
	Mod GetModMod `json:"mod"`
}

// GetMod returns GetModResponse.Mod, and is useful for accessing the field via an interface.
func (v *GetModResponse) GetMod() GetModMod { return v.Mod }

type ModFields string

const (
	ModFieldsCreatedAt       ModFields = "created_at"
	ModFieldsDownloads       ModFields = "downloads"
	ModFieldsHotness         ModFields = "hotness"
	ModFieldsLastVersionDate ModFields = "last_version_date"
	ModFieldsName            ModFields = "name"
	ModFieldsPopularity      ModFields = "popularity"
	ModFieldsSearch          ModFields = "search"
	ModFieldsUpdatedAt       ModFields = "updated_at"
	ModFieldsViews           ModFields = "views"
)

type ModFilter struct {
	Hidden     bool      `json:"hidden,omitempty"`
	Ids        []string  `json:"ids,omitempty"`
	Limit      int       `json:"limit,omitempty"`
	Offset     int       `json:"offset,omitempty"`
	Order      Order     `json:"order,omitempty"`
	Order_by   ModFields `json:"order_by,omitempty"`
	References []string  `json:"references,omitempty"`
	Search     string    `json:"search,omitempty"`
}

// GetHidden returns ModFilter.Hidden, and is useful for accessing the field via an interface.
func (v *ModFilter) GetHidden() bool { return v.Hidden }

// GetIds returns ModFilter.Ids, and is useful for accessing the field via an interface.
func (v *ModFilter) GetIds() []string { return v.Ids }

// GetLimit returns ModFilter.Limit, and is useful for accessing the field via an interface.
func (v *ModFilter) GetLimit() int { return v.Limit }

// GetOffset returns ModFilter.Offset, and is useful for accessing the field via an interface.
func (v *ModFilter) GetOffset() int { return v.Offset }

// GetOrder returns ModFilter.Order, and is useful for accessing the field via an interface.
func (v *ModFilter) GetOrder() Order { return v.Order }

// GetOrder_by returns ModFilter.Order_by, and is useful for accessing the field via an interface.
func (v *ModFilter) GetOrder_by() ModFields { return v.Order_by }

// GetReferences returns ModFilter.References, and is useful for accessing the field via an interface.
func (v *ModFilter) GetReferences() []string { return v.References }

// GetSearch returns ModFilter.Search, and is useful for accessing the field via an interface.
func (v *ModFilter) GetSearch() string { return v.Search }

type ModVersionConstraint struct {
	ModIdOrReference string `json:"modIdOrReference"`
	Version          string `json:"version"`
}

// GetModIdOrReference returns ModVersionConstraint.ModIdOrReference, and is useful for accessing the field via an interface.
func (v *ModVersionConstraint) GetModIdOrReference() string { return v.ModIdOrReference }

// GetVersion returns ModVersionConstraint.Version, and is useful for accessing the field via an interface.
func (v *ModVersionConstraint) GetVersion() string { return v.Version }

// ModVersionsMod includes the requested fields of the GraphQL type Mod.
type ModVersionsMod struct {
	Id       string                          `json:"id"`
	Versions []ModVersionsModVersionsVersion `json:"versions"`
}

// GetId returns ModVersionsMod.Id, and is useful for accessing the field via an interface.
func (v *ModVersionsMod) GetId() string { return v.Id }

// GetVersions returns ModVersionsMod.Versions, and is useful for accessing the field via an interface.
func (v *ModVersionsMod) GetVersions() []ModVersionsModVersionsVersion { return v.Versions }

// ModVersionsModVersionsVersion includes the requested fields of the GraphQL type Version.
type ModVersionsModVersionsVersion struct {
	Id      string `json:"id"`
	Version string `json:"version"`
}

// GetId returns ModVersionsModVersionsVersion.Id, and is useful for accessing the field via an interface.
func (v *ModVersionsModVersionsVersion) GetId() string { return v.Id }

// GetVersion returns ModVersionsModVersionsVersion.Version, and is useful for accessing the field via an interface.
func (v *ModVersionsModVersionsVersion) GetVersion() string { return v.Version }

// ModVersionsResponse is returned by ModVersions on success.
type ModVersionsResponse struct {
	Mod ModVersionsMod `json:"mod"`
}

// GetMod returns ModVersionsResponse.Mod, and is useful for accessing the field via an interface.
func (v *ModVersionsResponse) GetMod() ModVersionsMod { return v.Mod }

// ModsModsGetMods includes the requested fields of the GraphQL type GetMods.
type ModsModsGetMods struct {
	Count int                      `json:"count"`
	Mods  []ModsModsGetModsModsMod `json:"mods"`
}

// GetCount returns ModsModsGetMods.Count, and is useful for accessing the field via an interface.
func (v *ModsModsGetMods) GetCount() int { return v.Count }

// GetMods returns ModsModsGetMods.Mods, and is useful for accessing the field via an interface.
func (v *ModsModsGetMods) GetMods() []ModsModsGetModsModsMod { return v.Mods }

// ModsModsGetModsModsMod includes the requested fields of the GraphQL type Mod.
type ModsModsGetModsModsMod struct {
	Id                string    `json:"id"`
	Name              string    `json:"name"`
	Mod_reference     string    `json:"mod_reference"`
	Last_version_date time.Time `json:"-"`
	Created_at        time.Time `json:"-"`
	Views             int       `json:"views"`
	Downloads         int       `json:"downloads"`
	Popularity        int       `json:"popularity"`
	Hotness           int       `json:"hotness"`
}

// GetId returns ModsModsGetModsModsMod.Id, and is useful for accessing the field via an interface.
func (v *ModsModsGetModsModsMod) GetId() string { return v.Id }

// GetName returns ModsModsGetModsModsMod.Name, and is useful for accessing the field via an interface.
func (v *ModsModsGetModsModsMod) GetName() string { return v.Name }

// GetMod_reference returns ModsModsGetModsModsMod.Mod_reference, and is useful for accessing the field via an interface.
func (v *ModsModsGetModsModsMod) GetMod_reference() string { return v.Mod_reference }

// GetLast_version_date returns ModsModsGetModsModsMod.Last_version_date, and is useful for accessing the field via an interface.
func (v *ModsModsGetModsModsMod) GetLast_version_date() time.Time { return v.Last_version_date }

// GetCreated_at returns ModsModsGetModsModsMod.Created_at, and is useful for accessing the field via an interface.
func (v *ModsModsGetModsModsMod) GetCreated_at() time.Time { return v.Created_at }

// GetViews returns ModsModsGetModsModsMod.Views, and is useful for accessing the field via an interface.
func (v *ModsModsGetModsModsMod) GetViews() int { return v.Views }

// GetDownloads returns ModsModsGetModsModsMod.Downloads, and is useful for accessing the field via an interface.
func (v *ModsModsGetModsModsMod) GetDownloads() int { return v.Downloads }

// GetPopularity returns ModsModsGetModsModsMod.Popularity, and is useful for accessing the field via an interface.
func (v *ModsModsGetModsModsMod) GetPopularity() int { return v.Popularity }

// GetHotness returns ModsModsGetModsModsMod.Hotness, and is useful for accessing the field via an interface.
func (v *ModsModsGetModsModsMod) GetHotness() int { return v.Hotness }

func (v *ModsModsGetModsModsMod) UnmarshalJSON(b []byte) error {

	if string(b) == "null" {
		return nil
	}

	var firstPass struct {
		*ModsModsGetModsModsMod
		Last_version_date json.RawMessage `json:"last_version_date"`
		Created_at        json.RawMessage `json:"created_at"`
		graphql.NoUnmarshalJSON
	}
	firstPass.ModsModsGetModsModsMod = v

	err := json.Unmarshal(b, &firstPass)
	if err != nil {
		return err
	}

	{
		dst := &v.Last_version_date
		src := firstPass.Last_version_date
		if len(src) != 0 && string(src) != "null" {
			err = utils.UnmarshalDateTime(
				src, dst)
			if err != nil {
				return fmt.Errorf(
					"Unable to unmarshal ModsModsGetModsModsMod.Last_version_date: %w", err)
			}
		}
	}

	{
		dst := &v.Created_at
		src := firstPass.Created_at
		if len(src) != 0 && string(src) != "null" {
			err = utils.UnmarshalDateTime(
				src, dst)
			if err != nil {
				return fmt.Errorf(
					"Unable to unmarshal ModsModsGetModsModsMod.Created_at: %w", err)
			}
		}
	}
	return nil
}

type __premarshalModsModsGetModsModsMod struct {
	Id string `json:"id"`

	Name string `json:"name"`

	Mod_reference string `json:"mod_reference"`

	Last_version_date json.RawMessage `json:"last_version_date"`

	Created_at json.RawMessage `json:"created_at"`

	Views int `json:"views"`

	Downloads int `json:"downloads"`

	Popularity int `json:"popularity"`

	Hotness int `json:"hotness"`
}

func (v *ModsModsGetModsModsMod) MarshalJSON() ([]byte, error) {
	premarshaled, err := v.__premarshalJSON()
	if err != nil {
		return nil, err
	}
	return json.Marshal(premarshaled)
}

func (v *ModsModsGetModsModsMod) __premarshalJSON() (*__premarshalModsModsGetModsModsMod, error) {
	var retval __premarshalModsModsGetModsModsMod

	retval.Id = v.Id
	retval.Name = v.Name
	retval.Mod_reference = v.Mod_reference
	{

		dst := &retval.Last_version_date
		src := v.Last_version_date
		var err error
		*dst, err = json.Marshal(
			&src)
		if err != nil {
			return nil, fmt.Errorf(
				"Unable to marshal ModsModsGetModsModsMod.Last_version_date: %w", err)
		}
	}
	{

		dst := &retval.Created_at
		src := v.Created_at
		var err error
		*dst, err = json.Marshal(
			&src)
		if err != nil {
			return nil, fmt.Errorf(
				"Unable to marshal ModsModsGetModsModsMod.Created_at: %w", err)
		}
	}
	retval.Views = v.Views
	retval.Downloads = v.Downloads
	retval.Popularity = v.Popularity
	retval.Hotness = v.Hotness
	return &retval, nil
}

// ModsResponse is returned by Mods on success.
type ModsResponse struct {
	Mods ModsModsGetMods `json:"mods"`
}

// GetMods returns ModsResponse.Mods, and is useful for accessing the field via an interface.
func (v *ModsResponse) GetMods() ModsModsGetMods { return v.Mods }

type Order string

const (
	OrderAsc  Order = "asc"
	OrderDesc Order = "desc"
)

// ResolveModDependenciesModsModVersion includes the requested fields of the GraphQL type ModVersion.
type ResolveModDependenciesModsModVersion struct {
	Id            string                                                `json:"id"`
	Mod_reference string                                                `json:"mod_reference"`
	Versions      []ResolveModDependenciesModsModVersionVersionsVersion `json:"versions"`
}

// GetId returns ResolveModDependenciesModsModVersion.Id, and is useful for accessing the field via an interface.
func (v *ResolveModDependenciesModsModVersion) GetId() string { return v.Id }

// GetMod_reference returns ResolveModDependenciesModsModVersion.Mod_reference, and is useful for accessing the field via an interface.
func (v *ResolveModDependenciesModsModVersion) GetMod_reference() string { return v.Mod_reference }

// GetVersions returns ResolveModDependenciesModsModVersion.Versions, and is useful for accessing the field via an interface.
func (v *ResolveModDependenciesModsModVersion) GetVersions() []ResolveModDependenciesModsModVersionVersionsVersion {
	return v.Versions
}

// ResolveModDependenciesModsModVersionVersionsVersion includes the requested fields of the GraphQL type Version.
type ResolveModDependenciesModsModVersionVersionsVersion struct {
	Id           string                                                                             `json:"id"`
	Version      string                                                                             `json:"version"`
	Link         string                                                                             `json:"link"`
	Hash         string                                                                             `json:"hash"`
	Dependencies []ResolveModDependenciesModsModVersionVersionsVersionDependenciesVersionDependency `json:"dependencies"`
}

// GetId returns ResolveModDependenciesModsModVersionVersionsVersion.Id, and is useful for accessing the field via an interface.
func (v *ResolveModDependenciesModsModVersionVersionsVersion) GetId() string { return v.Id }

// GetVersion returns ResolveModDependenciesModsModVersionVersionsVersion.Version, and is useful for accessing the field via an interface.
func (v *ResolveModDependenciesModsModVersionVersionsVersion) GetVersion() string { return v.Version }

// GetLink returns ResolveModDependenciesModsModVersionVersionsVersion.Link, and is useful for accessing the field via an interface.
func (v *ResolveModDependenciesModsModVersionVersionsVersion) GetLink() string { return v.Link }

// GetHash returns ResolveModDependenciesModsModVersionVersionsVersion.Hash, and is useful for accessing the field via an interface.
func (v *ResolveModDependenciesModsModVersionVersionsVersion) GetHash() string { return v.Hash }

// GetDependencies returns ResolveModDependenciesModsModVersionVersionsVersion.Dependencies, and is useful for accessing the field via an interface.
func (v *ResolveModDependenciesModsModVersionVersionsVersion) GetDependencies() []ResolveModDependenciesModsModVersionVersionsVersionDependenciesVersionDependency {
	return v.Dependencies
}

// ResolveModDependenciesModsModVersionVersionsVersionDependenciesVersionDependency includes the requested fields of the GraphQL type VersionDependency.
type ResolveModDependenciesModsModVersionVersionsVersionDependenciesVersionDependency struct {
	Condition string `json:"condition"`
	Mod_id    string `json:"mod_id"`
	Optional  bool   `json:"optional"`
}

// GetCondition returns ResolveModDependenciesModsModVersionVersionsVersionDependenciesVersionDependency.Condition, and is useful for accessing the field via an interface.
func (v *ResolveModDependenciesModsModVersionVersionsVersionDependenciesVersionDependency) GetCondition() string {
	return v.Condition
}

// GetMod_id returns ResolveModDependenciesModsModVersionVersionsVersionDependenciesVersionDependency.Mod_id, and is useful for accessing the field via an interface.
func (v *ResolveModDependenciesModsModVersionVersionsVersionDependenciesVersionDependency) GetMod_id() string {
	return v.Mod_id
}

// GetOptional returns ResolveModDependenciesModsModVersionVersionsVersionDependenciesVersionDependency.Optional, and is useful for accessing the field via an interface.
func (v *ResolveModDependenciesModsModVersionVersionsVersionDependenciesVersionDependency) GetOptional() bool {
	return v.Optional
}

// ResolveModDependenciesResponse is returned by ResolveModDependencies on success.
type ResolveModDependenciesResponse struct {
	Mods []ResolveModDependenciesModsModVersion `json:"mods"`
}

// GetMods returns ResolveModDependenciesResponse.Mods, and is useful for accessing the field via an interface.
func (v *ResolveModDependenciesResponse) GetMods() []ResolveModDependenciesModsModVersion {
	return v.Mods
}

// SMLVersionsResponse is returned by SMLVersions on success.
type SMLVersionsResponse struct {
	SmlVersions SMLVersionsSmlVersionsGetSMLVersions `json:"smlVersions"`
}

// GetSmlVersions returns SMLVersionsResponse.SmlVersions, and is useful for accessing the field via an interface.
func (v *SMLVersionsResponse) GetSmlVersions() SMLVersionsSmlVersionsGetSMLVersions {
	return v.SmlVersions
}

// SMLVersionsSmlVersionsGetSMLVersions includes the requested fields of the GraphQL type GetSMLVersions.
type SMLVersionsSmlVersionsGetSMLVersions struct {
	Count        int                                                          `json:"count"`
	Sml_versions []SMLVersionsSmlVersionsGetSMLVersionsSml_versionsSMLVersion `json:"sml_versions"`
}

// GetCount returns SMLVersionsSmlVersionsGetSMLVersions.Count, and is useful for accessing the field via an interface.
func (v *SMLVersionsSmlVersionsGetSMLVersions) GetCount() int { return v.Count }

// GetSml_versions returns SMLVersionsSmlVersionsGetSMLVersions.Sml_versions, and is useful for accessing the field via an interface.
func (v *SMLVersionsSmlVersionsGetSMLVersions) GetSml_versions() []SMLVersionsSmlVersionsGetSMLVersionsSml_versionsSMLVersion {
	return v.Sml_versions
}

// SMLVersionsSmlVersionsGetSMLVersionsSml_versionsSMLVersion includes the requested fields of the GraphQL type SMLVersion.
type SMLVersionsSmlVersionsGetSMLVersionsSml_versionsSMLVersion struct {
	Id                   string `json:"id"`
	Version              string `json:"version"`
	Satisfactory_version int    `json:"satisfactory_version"`
}

// GetId returns SMLVersionsSmlVersionsGetSMLVersionsSml_versionsSMLVersion.Id, and is useful for accessing the field via an interface.
func (v *SMLVersionsSmlVersionsGetSMLVersionsSml_versionsSMLVersion) GetId() string { return v.Id }

// GetVersion returns SMLVersionsSmlVersionsGetSMLVersionsSml_versionsSMLVersion.Version, and is useful for accessing the field via an interface.
func (v *SMLVersionsSmlVersionsGetSMLVersionsSml_versionsSMLVersion) GetVersion() string {
	return v.Version
}

// GetSatisfactory_version returns SMLVersionsSmlVersionsGetSMLVersionsSml_versionsSMLVersion.Satisfactory_version, and is useful for accessing the field via an interface.
func (v *SMLVersionsSmlVersionsGetSMLVersionsSml_versionsSMLVersion) GetSatisfactory_version() int {
	return v.Satisfactory_version
}

type VersionFields string

const (
	VersionFieldsCreatedAt VersionFields = "created_at"
	VersionFieldsDownloads VersionFields = "downloads"
	VersionFieldsUpdatedAt VersionFields = "updated_at"
)

type VersionFilter struct {
	Ids      []string      `json:"ids,omitempty"`
	Limit    int           `json:"limit,omitempty"`
	Offset   int           `json:"offset,omitempty"`
	Order    Order         `json:"order,omitempty"`
	Order_by VersionFields `json:"order_by,omitempty"`
	Search   string        `json:"search,omitempty"`
}

// GetIds returns VersionFilter.Ids, and is useful for accessing the field via an interface.
func (v *VersionFilter) GetIds() []string { return v.Ids }

// GetLimit returns VersionFilter.Limit, and is useful for accessing the field via an interface.
func (v *VersionFilter) GetLimit() int { return v.Limit }

// GetOffset returns VersionFilter.Offset, and is useful for accessing the field via an interface.
func (v *VersionFilter) GetOffset() int { return v.Offset }

// GetOrder returns VersionFilter.Order, and is useful for accessing the field via an interface.
func (v *VersionFilter) GetOrder() Order { return v.Order }

// GetOrder_by returns VersionFilter.Order_by, and is useful for accessing the field via an interface.
func (v *VersionFilter) GetOrder_by() VersionFields { return v.Order_by }

// GetSearch returns VersionFilter.Search, and is useful for accessing the field via an interface.
func (v *VersionFilter) GetSearch() string { return v.Search }

// __GetModInput is used internally by genqlient
type __GetModInput struct {
	ModId string `json:"modId"`
}

// GetModId returns __GetModInput.ModId, and is useful for accessing the field via an interface.
func (v *__GetModInput) GetModId() string { return v.ModId }

// __ModVersionsInput is used internally by genqlient
type __ModVersionsInput struct {
	ModId  string        `json:"modId,omitempty"`
	Filter VersionFilter `json:"filter,omitempty"`
}

// GetModId returns __ModVersionsInput.ModId, and is useful for accessing the field via an interface.
func (v *__ModVersionsInput) GetModId() string { return v.ModId }

// GetFilter returns __ModVersionsInput.Filter, and is useful for accessing the field via an interface.
func (v *__ModVersionsInput) GetFilter() VersionFilter { return v.Filter }

// __ModsInput is used internally by genqlient
type __ModsInput struct {
	Filter ModFilter `json:"filter,omitempty"`
}

// GetFilter returns __ModsInput.Filter, and is useful for accessing the field via an interface.
func (v *__ModsInput) GetFilter() ModFilter { return v.Filter }

// __ResolveModDependenciesInput is used internally by genqlient
type __ResolveModDependenciesInput struct {
	Filter []ModVersionConstraint `json:"filter"`
}

// GetFilter returns __ResolveModDependenciesInput.Filter, and is useful for accessing the field via an interface.
func (v *__ResolveModDependenciesInput) GetFilter() []ModVersionConstraint { return v.Filter }

func GetMod(
	ctx context.Context,
	client graphql.Client,
	modId string,
) (*GetModResponse, error) {
	__input := __GetModInput{
		ModId: modId,
	}
	var err error

	var retval GetModResponse
	err = client.MakeRequest(
		ctx,
		"GetMod",
		`
query GetMod ($modId: String!) {
	mod: getModByIdOrReference(modIdOrReference: $modId) {
		id
		mod_reference
		name
		views
		downloads
		authors {
			role
			user {
				username
			}
		}
		full_description
		source_url
		created_at
	}
}
`,
		&retval,
		&__input,
	)
	return &retval, err
}

func ModVersions(
	ctx context.Context,
	client graphql.Client,
	modId string,
	filter VersionFilter,
) (*ModVersionsResponse, error) {
	__input := __ModVersionsInput{
		ModId:  modId,
		Filter: filter,
	}
	var err error

	var retval ModVersionsResponse
	err = client.MakeRequest(
		ctx,
		"ModVersions",
		`
query ModVersions ($modId: String!, $filter: VersionFilter) {
	mod: getModByIdOrReference(modIdOrReference: $modId) {
		id
		versions(filter: $filter) {
			id
			version
		}
	}
}
`,
		&retval,
		&__input,
	)
	return &retval, err
}

func Mods(
	ctx context.Context,
	client graphql.Client,
	filter ModFilter,
) (*ModsResponse, error) {
	__input := __ModsInput{
		Filter: filter,
	}
	var err error

	var retval ModsResponse
	err = client.MakeRequest(
		ctx,
		"Mods",
		`
query Mods ($filter: ModFilter) {
	mods: getMods(filter: $filter) {
		count
		mods {
			id
			name
			mod_reference
			last_version_date
			created_at
			views
			downloads
			popularity
			hotness
		}
	}
}
`,
		&retval,
		&__input,
	)
	return &retval, err
}

func ResolveModDependencies(
	ctx context.Context,
	client graphql.Client,
	filter []ModVersionConstraint,
) (*ResolveModDependenciesResponse, error) {
	__input := __ResolveModDependenciesInput{
		Filter: filter,
	}
	var err error

	var retval ResolveModDependenciesResponse
	err = client.MakeRequest(
		ctx,
		"ResolveModDependencies",
		`
query ResolveModDependencies ($filter: [ModVersionConstraint!]!) {
	mods: resolveModVersions(filter: $filter) {
		id
		mod_reference
		versions {
			id
			version
			link
			hash
			dependencies {
				condition
				mod_id
				optional
			}
		}
	}
}
`,
		&retval,
		&__input,
	)
	return &retval, err
}

func SMLVersions(
	ctx context.Context,
	client graphql.Client,
) (*SMLVersionsResponse, error) {
	var err error

	var retval SMLVersionsResponse
	err = client.MakeRequest(
		ctx,
		"SMLVersions",
		`
query SMLVersions {
	smlVersions: getSMLVersions(filter: {limit:100}) {
		count
		sml_versions {
			id
			version
			satisfactory_version
		}
	}
}
`,
		&retval,
		nil,
	)
	return &retval, err
}
