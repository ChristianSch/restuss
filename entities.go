package restuss

import "time"

// PersistedScan represents a Persisted Scan on Nessus API
type PersistedScan struct {
	ID                   int64  `json:"id"`
	UUID                 string `json:"uuid"`
	Name                 string `json:"name"`
	Status               string `json:"status"`
	CreationDate         int64  `json:"creation_date"`
	LastModificationDate int64  `json:"last_modification_date"`
	Owner                string `json:"owner"`
}

// Vulnerability represents a Vulnerability returned by Nessus API
type Vulnerability struct {
	VulnerabilityIndex int64  `json:"vuln_index"`
	Severity           int64  `json:"severity"`
	PluginName         string `json:"plugin_name"`
	Count              int64  `json:"count"`
	PluginID           int64  `json:"plugin_id"`
	PluginFamily       string `json:"plugin_family"`
}

// ScanDetail represents Details from a Scan returned by Nessus API
type ScanDetail struct {
	ID              int64
	Info            Info            `json:"info"`
	Hosts           []Host          `json:"hosts"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
}

// Info represents detailed information from a Scan returned by Nessus API
type Info struct {
	Status string `json:"status"`
}

// Host represents a host member of a scan
type Host struct {
	ID       int64  `json:"host_id"`
	Hostname string `json:"hostname"`
}

// ScanSettings represents settings for a Scan to be posted to Nessus API
type ScanSettings struct {
	Name     string `json:"name"`
	Enabled  bool   `json:"enabled"`
	Targets  string `json:"text_targets"`
	PolicyID int64  `json:"policy_id"`
}

// Scan represents a Scan to be posted to Nessus API
type Scan struct {
	TemplateUUID string       `json:"uuid"`
	Settings     ScanSettings `json:"settings"`
}

// ScanTemplate represents a Template for a Scan returned by Nessus API
type ScanTemplate struct {
	UUID             string `json:"uuid"`
	Name             string `json:"name"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	CloudOnly        bool   `json:"cloud_only"`
	SubscriptionOnly bool   `json:"subscription_only"`
	IsAgent          bool   `json:"is_agent"`
	Info             string `json:"more_info"`
}

// PluginAttribute represents an attribute for a plugin returned by Nessus API
type PluginAttribute struct {
	Name  string `json:"attribute_name"`
	Value string `json:"attribute_value"`
}

// Plugin represents a plugin returned by Nessus API
type Plugin struct {
	ID         int64             `json:"id"`
	Name       string            `json:"name"`
	FamilyName string            `json:"family_name"`
	Attributes []PluginAttribute `json:"attributes"`
}

// PluginOutputResponse represents the output of a plugin returned by Nessus API
type PluginOutputResponse struct {
	Output []PluginOutput `json:"outputs"`
}

// PluginOutput represents the output of a plugin returned by Nessus API
type PluginOutput struct {
	Hosts    string      `json:"hosts"`
	Ports    interface{} `json:"ports"`
	Output   string      `json:"plugin_output"`
	Severity int         `json:"severity"`
}

// Policy represents a policy returned by Nessus API
type Policy struct {
	ID       int64
	UUID     string         `json:"uuid"`
	Settings PolicySettings `json:"settings"`
}

// PolicySettings represents a setting for policy returned by Nessus API
type PolicySettings struct {
	Name string `json:"name"`
}

// Asset represents the asset entity on Tenable.io.
//
// More attributes available at
// https://developer.tenable.com/docs/common-asset-attributes.
type Asset struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	Types              []string  `json:"types"`
	Sources            []string  `json:"sources"`
	Created            time.Time `json:"created"`
	ObservationSources []struct {
		FirstObserved time.Time `json:"first_observed"`
		LastObserved  time.Time `json:"last_observed"`
		Name          string    `json:"name"`
	} `json:"observation_sources"`
	IsLicensed bool     `json:"is_licensed"`
	Fqdns      []string `json:"fqdns"`
	Tags       []struct {
		ID       string `json:"id"`
		Category string `json:"category"`
		Value    string `json:"value"`
		Type     string `json:"type"`
	} `json:"tags"`
	Network struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"network"`
	FirstObserved time.Time `json:"first_observed"`
	DisplayFqdn   string    `json:"display_fqdn"`
	IsDeleted     bool      `json:"is_deleted"`
	LastObserved  time.Time `json:"last_observed"`
	Updated       time.Time `json:"updated"`
}

// Finding represents the finding entity on Tenable.io.
//
// More info available at
// https://developer.tenable.com/docs/tenable-plugin-attributes.
type Finding struct {
	Output     string `json:"output"`
	ID         string `json:"id"`
	Severity   int    `json:"severity"`
	Port       int    `json:"port"`
	Protocol   string `json:"protocol"`
	Service    string `json:"service"`
	Definition struct {
		ID          int    `json:"id"` // plugin_id
		Name        string `json:"name"`
		Description string `json:"description"`
		Synopsis    string `json:"synopsis"`
		Solution    string `json"solution"`
		CVSS3       struct {
			BaseScore *float32 `json:"base_score"`
		} `json:"cvss3"`
		CVSS2 struct {
			BaseScore *float32 `json:"base_score"`
		} `json:"cvss2"`
		CWE     []string `json:"cwe"`
		SeeAlso []string `json:"see_also"`
	} `json:"definition"`
}

// Pagination is used to iterate results for some endpoints. If the attribute
// `Next` has content, needs to be passed as a parameter to the next request.
type Pagination struct {
	Next  string `json:"next"`
	Limit int    `json:"limit"`
	Total int    `json:"total"`
}
