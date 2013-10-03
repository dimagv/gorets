/**
	extraction of the data pieces describing a RETS system

	TODO - this class is ripe for removing redundancy
 */
package gorets

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Metadata struct {
	MSystem MSystem
	MResources MResources
	MClasses MClasses
	MTables MTables
	MLookups MLookups
	MLookupTypes MLookupTypes
}

type MetadataRequest struct {
	Url, Format, MType, Id string
}

func (s *Session) GetMetadata(r MetadataRequest) (*Metadata, error) {
	if r.MType == "*" {
		panic("not yet supported!")
	}

	// required
	values := url.Values{}
	values.Add("Format", r.Format)
	values.Add("Type", r.MType)
	values.Add("ID", r.Id)

	req, err := http.NewRequest(s.HttpMethod, r.Url+"?"+values.Encode(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.Client.Do(req)

	body,err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	metadata := Metadata{}

	// TOOD remove the needless repetition
	switch strings.ToUpper(r.Url) {
	case "METADATA-SYSTEM":
		tmp, err := parseMSystem(body)
		if err != nil {
			return nil, err
		}
		metadata.MSystem = *tmp
	case "METADATA-RESOURCE":
		tmp, err := parseMResources(body)
		if err != nil {
			return nil, err
		}
		metadata.MResources = *tmp
	case "METADATA-CLASS":
		tmp, err := parseMClasses(body)
		if err != nil {
			return nil, err
		}
		metadata.MClasses = *tmp
	case "METADATA-TABLE":
		tmp, err := parseMTables(body)
		if err != nil {
			return nil, err
		}
		metadata.MTables = *tmp
	case "METADATA-LOOKUP":
		tmp, err := parseMLookups(body)
		if err != nil {
			return nil, err
		}
		metadata.MLookups = *tmp
	case "METADATA-LOOKUP_TYPE":
		tmp, err := parseMLookupTypes(body)
		if err != nil {
			return nil, err
		}
		metadata.MLookupTypes = *tmp
	}


	return &metadata, nil
}

type MSystem struct {
	Version, Date string
	Id, Description string
	Comments string
}

func parseMSystem(response []byte) (*MSystem, error) {
	type XmlMSystem struct {
		Version string `xml:"Version,attr"`
		Date string `xml:"Date,attr"`
		Comments string `xml:"COMMENTS"`
	}
	type XmlSystem struct {
		SystemId string `xml:"SystemID,attr"`
		Description string `xml:"SystemDescription,attr"`
	}
	type XmlData struct {
		XMLName xml.Name `xml:"RETS"`
		ReplyCode int `xml:"ReplyCode,attr"`
		ReplyText string `xml:"ReplyText,attr"`
		MSystem XmlMSystem `xml:"METADATA-SYSTEM"`
		System XmlSystem `xml:"SYSTEM"`
	}

	decoder := xml.NewDecoder(bytes.NewBuffer(response))
	decoder.Strict = false

	xms := XmlData{}
	err := decoder.Decode(&xms)
	if err != nil {
		return nil, err
	}

	// transfer the contents to the public struct
	return &MSystem{
		Version: xms.MSystem.Version,
		Date: xms.MSystem.Date,
		Comments: strings.TrimSpace(xms.MSystem.Comments),
		Id: xms.System.SystemId,
		Description: xms.System.Description,
	}, nil
}


const delim = "	"

/* the common structure */
type MData struct {
	Version, Date string
	Columns []string
	Rows [][]string
}

/** cached lookup */
type Indexer func(col string, row int) string
/** create the cache */
func (m *MData) Indexer() Indexer {
	index := make(map[string]int)
	for i, c := range m.Columns {
		index[c] = i
	}
	return func(col string, row int) string {
		return m.Rows[row][index[col]]
	}
}

/** extract a map of fields from columns and rows */
func extractMap(cols string, rows []string) (*MData) {
	data := MData{}
	// remove the first and last chars
	data.Columns = strings.Split(strings.Trim(cols,delim),delim)
	data.Rows = make([][]string, len(rows))
	// create each
	for i,line := range rows {
		data.Rows[i] = strings.Split(strings.Trim(line,delim),delim)
	}
	return &data
}

type MResources struct {
	MData MData
}

func parseMResources(response []byte) (*MResources, error) {
	type XmlResource struct {
		Version string `xml:"Version,attr"`
		Date string `xml:"Date,attr"`
		Columns string `xml:"COLUMNS"`
		Data []string `xml:"DATA"`
	}
	type XmlData struct {
		XMLName xml.Name `xml:"RETS"`
		ReplyCode int `xml:"ReplyCode,attr"`
		ReplyText string `xml:"ReplyText,attr"`
		Info XmlResource `xml:"METADATA-RESOURCE"`
	}

	decoder := xml.NewDecoder(bytes.NewBuffer(response))
	decoder.Strict = false

	xms := XmlData{}
	err := decoder.Decode(&xms)
	if err != nil {
		return nil, err
	}

	// remove the first and last chars
	data := extractMap(xms.Info.Columns, xms.Info.Data)
	data.Date = xms.Info.Date
	data.Version = xms.Info.Version

	// transfer the contents to the public struct
	return &MResources{
		MData: *data,
	}, nil
}

type MClasses struct {
	MData MData
}

func parseMClasses(response []byte) (*MClasses, error) {
	type XmlClass struct {
		Resource string `xml:"Resource,attr"`
		Version string `xml:"Version,attr"`
		Date string `xml:"Date,attr"`
		Columns string `xml:"COLUMNS"`
		Data []string `xml:"DATA"`
	}
	type XmlData struct {
		XMLName xml.Name `xml:"RETS"`
		ReplyCode int `xml:"ReplyCode,attr"`
		ReplyText string `xml:"ReplyText,attr"`
		Info XmlClass `xml:"METADATA-CLASS"`
	}

	decoder := xml.NewDecoder(bytes.NewBuffer(response))
	decoder.Strict = false

	xms := XmlData{}
	err := decoder.Decode(&xms)
	if err != nil {
		return nil, err
	}

	// remove the first and last chars
	data := extractMap(xms.Info.Columns, xms.Info.Data)
	data.Date = xms.Info.Date
	data.Version = xms.Info.Version

	// transfer the contents to the public struct
	return &MClasses{
		MData: *data,
	}, nil
}

type MTables struct {
	MData MData
}

func parseMTables(response []byte) (*MTables, error) {
	type XmlTable struct {
		Resource string `xml:"Resource,attr"`
		Version string `xml:"Version,attr"`
		Date string `xml:"Date,attr"`
		Columns string `xml:"COLUMNS"`
		Data []string `xml:"DATA"`
	}
	type XmlData struct {
		XMLName xml.Name `xml:"RETS"`
		ReplyCode int `xml:"ReplyCode,attr"`
		ReplyText string `xml:"ReplyText,attr"`
		Info XmlTable `xml:"METADATA-TABLE"`
	}

	decoder := xml.NewDecoder(bytes.NewBuffer(response))
	decoder.Strict = false

	xms := XmlData{}
	err := decoder.Decode(&xms)
	if err != nil {
		return nil, err
	}

	// remove the first and last chars
	data := extractMap(xms.Info.Columns, xms.Info.Data)
	data.Date = xms.Info.Date
	data.Version = xms.Info.Version

	// transfer the contents to the public struct
	return &MTables{
		MData: *data,
	}, nil
}

type MLookups struct {
	MData MData
}

func parseMLookups(response []byte) (*MLookups, error) {
	type XmlTable struct {
		Resource string `xml:"Resource,attr"`
		Version string `xml:"Version,attr"`
		Date string `xml:"Date,attr"`
		Columns string `xml:"COLUMNS"`
		Data []string `xml:"DATA"`
	}
	type XmlData struct {
		XMLName xml.Name `xml:"RETS"`
		ReplyCode int `xml:"ReplyCode,attr"`
		ReplyText string `xml:"ReplyText,attr"`
		Info XmlTable `xml:"METADATA-LOOKUP"`
	}

	decoder := xml.NewDecoder(bytes.NewBuffer(response))
	decoder.Strict = false

	xms := XmlData{}
	err := decoder.Decode(&xms)
	if err != nil {
		return nil, err
	}

	// remove the first and last chars
	data := extractMap(xms.Info.Columns, xms.Info.Data)
	data.Date = xms.Info.Date
	data.Version = xms.Info.Version

	// transfer the contents to the public struct
	return &MLookups{
		MData: *data,
	}, nil
}

type MLookupTypes struct {
	MData MData
}

func parseMLookupTypes(response []byte) (*MLookupTypes, error) {
	type XmlTable struct {
		Resource string `xml:"Resource,attr"`
		Version string `xml:"Version,attr"`
		Date string `xml:"Date,attr"`
		Columns string `xml:"COLUMNS"`
		Data []string `xml:"DATA"`
	}
	type XmlData struct {
		XMLName xml.Name `xml:"RETS"`
		ReplyCode int `xml:"ReplyCode,attr"`
		ReplyText string `xml:"ReplyText,attr"`
		Info XmlTable `xml:"METADATA-LOOKUP_TYPE"`
	}

	decoder := xml.NewDecoder(bytes.NewBuffer(response))
	decoder.Strict = false

	xms := XmlData{}
	err := decoder.Decode(&xms)
	if err != nil {
		return nil, err
	}

	// remove the first and last chars
	data := extractMap(xms.Info.Columns, xms.Info.Data)
	data.Date = xms.Info.Date
	data.Version = xms.Info.Version

	// transfer the contents to the public struct
	return &MLookupTypes{
		MData: *data,
	}, nil
}
