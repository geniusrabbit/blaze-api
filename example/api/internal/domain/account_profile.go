package domain

import "github.com/geniusrabbit/gosql/v2"

// AccountProfile is example/api profile metadata (consumer-defined per project).
type AccountProfile struct {
	Title       string `json:"title"`
	Description string `json:"description"`

	LogoURI string `json:"logo_uri" gorm:"column:logo_uri"`

	PolicyURI         string `json:"policy_uri" gorm:"column:policy_uri"`
	TermsOfServiceURI string `json:"tos_uri" gorm:"column:tos_uri"`
	ClientURI         string `json:"client_uri" gorm:"column:client_uri"`

	Contacts gosql.NullableStringArray `json:"contacts" gorm:"column:contacts;type:text[]"`
}

// GetTitle returns account title.
func (p *AccountProfile) GetTitle() string {
	if p == nil {
		return ""
	}
	return p.Title
}

// GetDescription returns account description.
func (p *AccountProfile) GetDescription() string {
	if p == nil {
		return ""
	}
	return p.Description
}

// GetLogoURI returns logo URI.
func (p *AccountProfile) GetLogoURI() string {
	if p == nil {
		return ""
	}
	return p.LogoURI
}

// GetPolicyURI returns policy URI.
func (p *AccountProfile) GetPolicyURI() string {
	if p == nil {
		return ""
	}
	return p.PolicyURI
}

// GetTermsOfServiceURI returns terms of service URI.
func (p *AccountProfile) GetTermsOfServiceURI() string {
	if p == nil {
		return ""
	}
	return p.TermsOfServiceURI
}

// GetClientURI returns client URI.
func (p *AccountProfile) GetClientURI() string {
	if p == nil {
		return ""
	}
	return p.ClientURI
}

// GetContacts returns contact list.
func (p *AccountProfile) GetContacts() []string {
	if p == nil {
		return nil
	}
	return append([]string(nil), p.Contacts...)
}

// ApplyProfile sets profile fields from input values.
func (p *AccountProfile) ApplyProfile(title, description, logoURI, policyURI, tosURI, clientURI string, contacts []string) {
	if p == nil {
		return
	}
	p.Title = title
	p.Description = description
	p.LogoURI = logoURI
	p.PolicyURI = policyURI
	p.TermsOfServiceURI = tosURI
	p.ClientURI = clientURI
	p.Contacts = contacts
}
