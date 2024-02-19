// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package models

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/geniusrabbit/blaze-api/server/graphql/types"
	"github.com/google/uuid"
)

// Account is a company account that can be used to login to the system.
type Account struct {
	// The primary key of the Account
	ID uint64 `json:"ID"`
	// Status of Account active
	Status ApproveStatus `json:"status"`
	// Message which defined during user approve/rejection process
	StatusMessage *string `json:"statusMessage,omitempty"`
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	// logoURI is an URL string that references a logo for the client.
	LogoURI string `json:"logoURI"`
	// policyURI is a URL string that points to a human-readable privacy policy document
	// that describes how the deployment organization collects, uses,
	// retains, and discloses personal data.
	PolicyURI string `json:"policyURI"`
	// termsOfServiceURI is a URL string that points to a human-readable terms of service
	// document for the client that describes a contractual relationship
	// between the end-user and the client that the end-user accepts when
	// authorizing the client.
	TermsOfServiceURI string `json:"termsOfServiceURI"`
	// clientURI is an URL string of a web page providing information about the client.
	// If present, the server SHOULD display this URL to the end-user in
	// a clickable fashion.
	ClientURI string `json:"clientURI"`
	// contacts is a array of strings representing ways to contact people responsible
	// for this client, typically email addresses.
	Contacts  []string  `json:"contacts,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type AccountCreateInput struct {
	OwnerID  *uint64       `json:"ownerID,omitempty"`
	Owner    *UserInput    `json:"owner,omitempty"`
	Account  *AccountInput `json:"account"`
	Password string        `json:"password"`
}

type AccountCreatePayload struct {
	// A unique identifier for the client performing the mutation.
	ClientMutationID string `json:"clientMutationID"`
	// The account object
	Account *Account `json:"account"`
	// The user object
	Owner *User `json:"owner"`
}

type AccountEdge struct {
	// A cursor for use in pagination.
	Cursor string `json:"cursor"`
	// The item at the end of the edge.
	Node *Account `json:"node,omitempty"`
}

type AccountInput struct {
	Status            *ApproveStatus `json:"status,omitempty"`
	Title             *string        `json:"title,omitempty"`
	Description       *string        `json:"description,omitempty"`
	LogoURI           *string        `json:"logoURI,omitempty"`
	PolicyURI         *string        `json:"policyURI,omitempty"`
	TermsOfServiceURI *string        `json:"termsOfServiceURI,omitempty"`
	ClientURI         *string        `json:"clientURI,omitempty"`
	Contacts          []string       `json:"contacts,omitempty"`
}

type AccountListFilter struct {
	ID     []uint64        `json:"ID,omitempty"`
	UserID []uint64        `json:"UserID,omitempty"`
	Title  []string        `json:"title,omitempty"`
	Status []ApproveStatus `json:"status,omitempty"`
}

type AccountListOrder struct {
	ID     *Ordering `json:"ID,omitempty"`
	Title  *Ordering `json:"title,omitempty"`
	Status *Ordering `json:"status,omitempty"`
}

// AccountPayload wrapper to access of Account oprtation results
type AccountPayload struct {
	// A unique identifier for the client performing the mutation.
	ClientMutationID string `json:"clientMutationID"`
	// Account ID operation result
	AccountID uint64 `json:"accountID"`
	// Account object accessor
	Account *Account `json:"account,omitempty"`
}

// AuthClient object represents an OAuth 2.0 client
type AuthClient struct {
	// ClientID is the client ID which represents unique connection indentificator
	ID        string `json:"ID"`
	AccountID uint64 `json:"accountID"`
	UserID    uint64 `json:"userID"`
	// Title of the AuthClient as himan readable name
	Title string `json:"title"`
	// Secret is the client's secret. The secret will be included in the create request as cleartext, and then
	// never again. The secret is stored using BCrypt so it is impossible to recover it. Tell your users
	// that they need to write the secret down as it will not be made available again.
	Secret string `json:"secret"`
	// RedirectURIs is an array of allowed redirect urls for the client, for example http://mydomain/oauth/callback .
	RedirectURIs []string `json:"redirectURIs,omitempty"`
	// GrantTypes is an array of grant types the client is allowed to use.
	//
	// Pattern: client_credentials|authorization_code|implicit|refresh_token
	GrantTypes []string `json:"grantTypes,omitempty"`
	// ResponseTypes is an array of the OAuth 2.0 response type strings that the client can
	// use at the authorization endpoint.
	//
	// Pattern: id_token|code|token
	ResponseTypes []string `json:"responseTypes,omitempty"`
	// Scope is a string containing a space-separated list of scope values (as
	// described in Section 3.3 of OAuth 2.0 [RFC6749]) that the client
	// can use when requesting access tokens.
	//
	// Pattern: ([a-zA-Z0-9\.\*]+\s?)+
	Scope string `json:"scope"`
	// Audience is a whitelist defining the audiences this client is allowed to request tokens for. An audience limits
	// the applicability of an OAuth 2.0 Access Token to, for example, certain API endpoints. The value is a list
	// of URLs. URLs MUST NOT contain whitespaces.
	Audience []string `json:"audience,omitempty"`
	// SubjectType requested for responses to this Client. The subject_types_supported Discovery parameter contains a
	// list of the supported subject_type values for this server. Valid types include `pairwise` and `public`.
	SubjectType string `json:"subjectType"`
	// AllowedCORSOrigins are one or more URLs (scheme://host[:port]) which are allowed to make CORS requests
	// to the /oauth/token endpoint. If this array is empty, the sever's CORS origin configuration (`CORS_ALLOWED_ORIGINS`)
	// will be used instead. If this array is set, the allowed origins are appended to the server's CORS origin configuration.
	// Be aware that environment variable `CORS_ENABLED` MUST be set to `true` for this to work.
	AllowedCORSOrigins []string `json:"allowedCORSOrigins,omitempty"`
	// Public flag tells that the client is public
	Public bool `json:"public"`
	// ExpiresAt contins the time of expiration of the client
	ExpiresAt time.Time  `json:"expiresAt"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}

type AuthClientEdge struct {
	// A cursor for use in pagination.
	Cursor string `json:"cursor"`
	// The item at the end of the edge.
	Node *AuthClient `json:"node,omitempty"`
}

type AuthClientInput struct {
	AccountID          *uint64    `json:"accountID,omitempty"`
	UserID             *uint64    `json:"userID,omitempty"`
	Title              *string    `json:"title,omitempty"`
	Secret             *string    `json:"secret,omitempty"`
	RedirectURIs       []string   `json:"redirectURIs,omitempty"`
	GrantTypes         []string   `json:"grantTypes,omitempty"`
	ResponseTypes      []string   `json:"responseTypes,omitempty"`
	Scope              *string    `json:"scope,omitempty"`
	Audience           []string   `json:"audience,omitempty"`
	SubjectType        string     `json:"subjectType"`
	AllowedCORSOrigins []string   `json:"allowedCORSOrigins,omitempty"`
	Public             *bool      `json:"public,omitempty"`
	ExpiresAt          *time.Time `json:"expiresAt,omitempty"`
}

type AuthClientListFilter struct {
	ID        []string `json:"ID,omitempty"`
	UserID    []uint64 `json:"userID,omitempty"`
	AccountID []uint64 `json:"accountID,omitempty"`
	Public    *bool    `json:"public,omitempty"`
}

type AuthClientListOrder struct {
	ID         *Ordering `json:"ID,omitempty"`
	UserID     *Ordering `json:"userID,omitempty"`
	AccountID  *Ordering `json:"accountID,omitempty"`
	Title      *Ordering `json:"title,omitempty"`
	Public     *Ordering `json:"public,omitempty"`
	LastUpdate *Ordering `json:"lastUpdate,omitempty"`
}

// AuthClientPayload wrapper to access of AuthClient oprtation results
type AuthClientPayload struct {
	// A unique identifier for the client performing the mutation.
	ClientMutationID string `json:"clientMutationID"`
	// AuthClient ID operation result
	AuthClientID string `json:"authClientID"`
	// AuthClient object accessor
	AuthClient *AuthClient `json:"authClient,omitempty"`
}

// HistoryAction is the model for history actions.
type HistoryAction struct {
	ID         uuid.UUID          `json:"ID"`
	Name       string             `json:"name"`
	Message    string             `json:"message"`
	UserID     uint64             `json:"userID"`
	AccountID  uint64             `json:"accountID"`
	ObjectType string             `json:"objectType"`
	ObjectID   uint64             `json:"objectID"`
	ObjectIDs  string             `json:"objectIDs"`
	Data       types.NullableJSON `json:"data"`
	ActionAt   time.Time          `json:"actionAt"`
}

// Edge of action history object.
type HistoryActionEdge struct {
	// The item at the end of the edge.
	Node *HistoryAction `json:"node"`
	// A cursor for use in pagination.
	Cursor string `json:"cursor"`
}

type HistoryActionListFilter struct {
	ID []uuid.UUID `json:"ID,omitempty"`
	// The name of the action
	Name []string `json:"name,omitempty"`
	// List of users who made the action
	UserID []uint64 `json:"userID,omitempty"`
	// List of accounts that the user belongs to
	AccountID []uint64 `json:"accountID,omitempty"`
	// Type of the object that the action is performed on
	ObjectType []string `json:"objectType,omitempty"`
	// Object ID of the model that the action is performed on
	ObjectID []uint64 `json:"objectID,omitempty"`
	// Object ID string version of the model that the action is performed on
	ObjectIDs []string `json:"objectIDs,omitempty"`
}

// HistoryActionListOptions contains the options for listing history actions ordering.
type HistoryActionListOrder struct {
	ID         *Ordering `json:"ID,omitempty"`
	Name       *Ordering `json:"name,omitempty"`
	UserID     *Ordering `json:"userID,omitempty"`
	AccountID  *Ordering `json:"accountID,omitempty"`
	ObjectType *Ordering `json:"objectType,omitempty"`
	ObjectID   *Ordering `json:"objectID,omitempty"`
	ObjectIDs  *Ordering `json:"objectIDs,omitempty"`
	ActionAt   *Ordering `json:"actionAt,omitempty"`
}

// HistoryActionPayload contains the information about a history action.
type HistoryActionPayload struct {
	// The client mutation id
	ClientMutationID *string `json:"clientMutationId,omitempty"`
	// The history action object ID
	ActionID uuid.UUID `json:"actionID"`
	// The action object
	Action *HistoryAction `json:"action"`
}

type Mutation struct {
}

// Option type definition represents a single option of the user or the system.
type Option struct {
	OptionType OptionType          `json:"optionType"`
	TargetID   uint64              `json:"targetID"`
	Name       string              `json:"name"`
	Value      *types.NullableJSON `json:"value,omitempty"`
}

// The edge type for Option.
type OptionEdge struct {
	Cursor string  `json:"cursor"`
	Node   *Option `json:"node"`
}

type OptionInput struct {
	// The type of the option.
	OptionType OptionType `json:"optionType"`
	// The target ID of the option.
	TargetID uint64 `json:"targetID"`
	// Value of the option.
	Value *types.NullableJSON `json:"value,omitempty"`
}

type OptionListFilter struct {
	OptionType  []OptionType `json:"optionType,omitempty"`
	TargetID    []uint64     `json:"targetID,omitempty"`
	Name        []string     `json:"name,omitempty"`
	NamePattern []string     `json:"namePattern,omitempty"`
}

type OptionListOrder struct {
	OptionType *Ordering `json:"optionType,omitempty"`
	TargetID   *Ordering `json:"targetID,omitempty"`
	Name       *Ordering `json:"name,omitempty"`
	Value      *Ordering `json:"value,omitempty"`
}

type OptionPayload struct {
	// A unique identifier for the client performing the mutation.
	ClientMutationID string  `json:"clientMutationId"`
	OptionName       string  `json:"optionName"`
	Option           *Option `json:"option,omitempty"`
}

// Information for paginating
type Page struct {
	// Start after the cursor ID
	After *string `json:"after,omitempty"`
	// Page number to start at (0-based), defaults to 0 (0, 1, 2, etc.)
	StartPage *int `json:"startPage,omitempty"`
	// Maximum number of items to return
	Size *int `json:"size,omitempty"`
}

// Information for paginating
type PageInfo struct {
	// When paginating backwards, the cursor to continue.
	StartCursor string `json:"startCursor"`
	// When paginating forwards, the cursor to continue.
	EndCursor string `json:"endCursor"`
	// When paginating backwards, are there more items?
	HasPreviousPage bool `json:"hasPreviousPage"`
	// When paginating forwards, are there more items?
	HasNextPage bool `json:"hasNextPage"`
	// Total number of pages available
	Total int `json:"total"`
	// Current page number
	Page int `json:"page"`
	// Number of pages
	Count int `json:"count"`
}

type Profile struct {
	ID          uint64              `json:"ID"`
	User        *User               `json:"user"`
	FirstName   string              `json:"firstName"`
	LastName    string              `json:"lastName"`
	CompanyName string              `json:"companyName"`
	About       string              `json:"about"`
	Email       string              `json:"email"`
	Messgangers []*ProfileMessanger `json:"messgangers,omitempty"`
	CreatedAt   time.Time           `json:"createdAt"`
	UpdatedAt   time.Time           `json:"updatedAt"`
}

type ProfileMessanger struct {
	Mtype   MessangerType `json:"mtype"`
	Address string        `json:"address"`
}

type Query struct {
}

type RBACPermission struct {
	Name   string `json:"name"`
	Object string `json:"object"`
	Access string `json:"access"`
}

// A role is a collection of permissions. A role can be a child of another role.
type RBACRole struct {
	ID    uint64 `json:"ID"`
	Name  string `json:"name"`
	Title string `json:"title"`
	//  Context is a JSON object that defines the context of the role.
	//  The context is used to determine whether the role is applicable to the object.
	//  The context is a JSON object with the following structure:
	//
	// {"cover": "system", "object": "role"}
	//
	//  where:
	// "cover" - is a name of the cover area of the object type
	// "object" - is a name of the object type <module>:<object-name>
	Context            *types.NullableJSON `json:"context,omitempty"`
	ChildRoles         []*RBACRole         `json:"childRoles,omitempty"`
	Permissions        []*RBACPermission   `json:"permissions,omitempty"`
	PermissionPatterns []string            `json:"permissionPatterns,omitempty"`
	CreatedAt          time.Time           `json:"createdAt"`
	UpdatedAt          time.Time           `json:"updatedAt"`
	DeletedAt          *time.Time          `json:"deletedAt,omitempty"`
}

// RBACRoleEdge is a connection edge type for RBACRole.
type RBACRoleEdge struct {
	// A cursor for use in pagination.
	Cursor string `json:"cursor"`
	// The item at the end of the edge.
	Node *RBACRole `json:"node,omitempty"`
}

type RBACRoleInput struct {
	Name        *string             `json:"name,omitempty"`
	Title       *string             `json:"title,omitempty"`
	Context     *types.NullableJSON `json:"context,omitempty"`
	Permissions []string            `json:"permissions,omitempty"`
}

type RBACRoleListFilter struct {
	ID   []uint64 `json:"ID,omitempty"`
	Name []string `json:"name,omitempty"`
}

type RBACRoleListOrder struct {
	ID    *Ordering `json:"ID,omitempty"`
	Name  *Ordering `json:"name,omitempty"`
	Title *Ordering `json:"title,omitempty"`
}

// RBACRolePayload wrapper to access of RBACRole oprtation results
type RBACRolePayload struct {
	// A unique identifier for the client performing the mutation.
	ClientMutationID string `json:"clientMutationID"`
	// Role ID operation result
	RoleID uint64 `json:"roleID"`
	// Role object accessor
	Role *RBACRole `json:"role,omitempty"`
}

// SessionToken object represents an OAuth 2.0 / JWT session token
type SessionToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	IsAdmin   bool      `json:"isAdmin"`
	Roles     []string  `json:"roles,omitempty"`
}

// Simple response type for the API
type StatusResponse struct {
	// The status of the response
	Status ResponseStatus `json:"status"`
	// The message of the response
	Message *string `json:"message,omitempty"`
}

// User represents a user object of the system
type User struct {
	// The primary key of the user
	ID uint64 `json:"ID"`
	// Unical user name
	Username string `json:"username"`
	// Status of user active
	Status ApproveStatus `json:"status"`
	// Message which defined during user approve/rejection process
	StatusMessage *string   `json:"statusMessage,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type UserEdge struct {
	// A cursor for use in pagination.
	Cursor string `json:"cursor"`
	// The item at the end of the edge.
	Node *User `json:"node,omitempty"`
}

type UserInput struct {
	Username *string        `json:"username,omitempty"`
	Status   *ApproveStatus `json:"status,omitempty"`
}

// UserListFilter implements filter for user list query
type UserListFilter struct {
	ID        []uint64 `json:"ID,omitempty"`
	AccountID []uint64 `json:"accountID,omitempty"`
	Emails    []string `json:"emails,omitempty"`
	Roles     []uint64 `json:"roles,omitempty"`
}

// UserListOrder implements order for user list query
type UserListOrder struct {
	ID               *Ordering `json:"ID,omitempty"`
	Email            *Ordering `json:"email,omitempty"`
	Username         *Ordering `json:"username,omitempty"`
	Status           *Ordering `json:"status,omitempty"`
	RegistrationDate *Ordering `json:"registrationDate,omitempty"`
	Country          *Ordering `json:"country,omitempty"`
	Manager          *Ordering `json:"manager,omitempty"`
	CreatedAt        *Ordering `json:"createdAt,omitempty"`
	UpdatedAt        *Ordering `json:"updatedAt,omitempty"`
}

// UserPayload wrapper to access of user oprtation results
type UserPayload struct {
	// A unique identifier for the client performing the mutation.
	ClientMutationID string `json:"clientMutationID"`
	// User ID operation result
	UserID uint64 `json:"userID"`
	// User object accessor
	User *User `json:"user,omitempty"`
}

// The list of statuses that shows is particular object active or paused
type ActiveStatus string

const (
	// All object by default have to be paused
	ActiveStatusPaused ActiveStatus = "PAUSED"
	// Status of the active object
	ActiveStatusActive ActiveStatus = "ACTIVE"
)

var AllActiveStatus = []ActiveStatus{
	ActiveStatusPaused,
	ActiveStatusActive,
}

func (e ActiveStatus) IsValid() bool {
	switch e {
	case ActiveStatusPaused, ActiveStatusActive:
		return true
	}
	return false
}

func (e ActiveStatus) String() string {
	return string(e)
}

func (e *ActiveStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ActiveStatus(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ActiveStatus", str)
	}
	return nil
}

func (e ActiveStatus) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

// The list of statuses that shows is object approved or not
type ApproveStatus string

const (
	// Pending status of the just inited objects
	ApproveStatusPending ApproveStatus = "PENDING"
	// Approved status of object could be obtained from the some authorized user who have permissions
	ApproveStatusApproved ApproveStatus = "APPROVED"
	// Rejected status of object could be obtained from the some authorized user who have permissions
	ApproveStatusRejected ApproveStatus = "REJECTED"
)

var AllApproveStatus = []ApproveStatus{
	ApproveStatusPending,
	ApproveStatusApproved,
	ApproveStatusRejected,
}

func (e ApproveStatus) IsValid() bool {
	switch e {
	case ApproveStatusPending, ApproveStatusApproved, ApproveStatusRejected:
		return true
	}
	return false
}

func (e ApproveStatus) String() string {
	return string(e)
}

func (e *ApproveStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ApproveStatus(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ApproveStatus", str)
	}
	return nil
}

func (e ApproveStatus) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

// The list of statuses that shows is particular object is available
type AvailableStatus string

const (
	// All object by default have to be undefined
	AvailableStatusUndefined AvailableStatus = "UNDEFINED"
	// Status of the available object
	AvailableStatusAvailable AvailableStatus = "AVAILABLE"
	// Status of the unavailable object
	AvailableStatusUnavailable AvailableStatus = "UNAVAILABLE"
)

var AllAvailableStatus = []AvailableStatus{
	AvailableStatusUndefined,
	AvailableStatusAvailable,
	AvailableStatusUnavailable,
}

func (e AvailableStatus) IsValid() bool {
	switch e {
	case AvailableStatusUndefined, AvailableStatusAvailable, AvailableStatusUnavailable:
		return true
	}
	return false
}

func (e AvailableStatus) String() string {
	return string(e)
}

func (e *AvailableStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = AvailableStatus(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid AvailableStatus", str)
	}
	return nil
}

func (e AvailableStatus) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type MessangerType string

const (
	MessangerTypeSkype    MessangerType = "SKYPE"
	MessangerTypeAim      MessangerType = "AIM"
	MessangerTypeIcq      MessangerType = "ICQ"
	MessangerTypeWhatsapp MessangerType = "WHATSAPP"
	MessangerTypeTelegram MessangerType = "TELEGRAM"
	MessangerTypeViber    MessangerType = "VIBER"
	MessangerTypePhone    MessangerType = "PHONE"
)

var AllMessangerType = []MessangerType{
	MessangerTypeSkype,
	MessangerTypeAim,
	MessangerTypeIcq,
	MessangerTypeWhatsapp,
	MessangerTypeTelegram,
	MessangerTypeViber,
	MessangerTypePhone,
}

func (e MessangerType) IsValid() bool {
	switch e {
	case MessangerTypeSkype, MessangerTypeAim, MessangerTypeIcq, MessangerTypeWhatsapp, MessangerTypeTelegram, MessangerTypeViber, MessangerTypePhone:
		return true
	}
	return false
}

func (e MessangerType) String() string {
	return string(e)
}

func (e *MessangerType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = MessangerType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid MessangerType", str)
	}
	return nil
}

func (e MessangerType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type OptionType string

const (
	OptionTypeUndefined OptionType = "UNDEFINED"
	OptionTypeUser      OptionType = "USER"
	OptionTypeAccount   OptionType = "ACCOUNT"
	OptionTypeSystem    OptionType = "SYSTEM"
)

var AllOptionType = []OptionType{
	OptionTypeUndefined,
	OptionTypeUser,
	OptionTypeAccount,
	OptionTypeSystem,
}

func (e OptionType) IsValid() bool {
	switch e {
	case OptionTypeUndefined, OptionTypeUser, OptionTypeAccount, OptionTypeSystem:
		return true
	}
	return false
}

func (e OptionType) String() string {
	return string(e)
}

func (e *OptionType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = OptionType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid OptionType", str)
	}
	return nil
}

func (e OptionType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

// Constants of the order of data
type Ordering string

const (
	// Ascending ordering of data
	OrderingAsc Ordering = "ASC"
	// Descending ordering of data
	OrderingDesc Ordering = "DESC"
)

var AllOrdering = []Ordering{
	OrderingAsc,
	OrderingDesc,
}

func (e Ordering) IsValid() bool {
	switch e {
	case OrderingAsc, OrderingDesc:
		return true
	}
	return false
}

func (e Ordering) String() string {
	return string(e)
}

func (e *Ordering) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Ordering(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Ordering", str)
	}
	return nil
}

func (e Ordering) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

// Constants of the response status
type ResponseStatus string

const (
	// Success status of the response
	ResponseStatusSuccess ResponseStatus = "SUCCESS"
	// Error status of the response
	ResponseStatusError ResponseStatus = "ERROR"
)

var AllResponseStatus = []ResponseStatus{
	ResponseStatusSuccess,
	ResponseStatusError,
}

func (e ResponseStatus) IsValid() bool {
	switch e {
	case ResponseStatusSuccess, ResponseStatusError:
		return true
	}
	return false
}

func (e ResponseStatus) String() string {
	return string(e)
}

func (e *ResponseStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ResponseStatus(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ResponseStatus", str)
	}
	return nil
}

func (e ResponseStatus) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
