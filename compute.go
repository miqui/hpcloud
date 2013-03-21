package hpcloud

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
)

type Flavor int

const (
	XSmall = Flavor(100) + iota
	Small
	Medium
	Large
	XLarge
	DblXLarge
)

type ServerImage int

const (
	UbuntuLucid10_04Kernel    = ServerImage(1235)
	UbuntuLucid10_04          = 1236
	UbuntuMaverick10_10Kernel = 1237
	UbuntuMaverick10_10       = 1238
	UbuntuNatty11_04Kernel    = 1239
	UbuntuNatty11_04          = 1240
	UbuntuOneiric11_10        = 5579
	UbuntuPrecise12_04        = 8419
	CentOS5_8Server64         = 54021
	CentOS6_2Server64Kernel   = 1356
	CentOS6_2Server64Ramdisk  = 1357
	CentOS6_2Server64         = 1358
	DebianSqueeze6_0_3Kernel  = 1359
	DebianSqueeze6_0_3Ramdisk = 1360
	DebianSqueeze6_0_3Server  = 1361
	Fedora16Server64          = 16291
	BitNamiDrupal7_14_0       = 22729
	BitNamiWebPack1_2_0       = 22731
	BitNamiDevPack1_0_0       = 4654
	ActiveStateStackatov1_2_6 = 14345
	ActiveStateStackatov2_2_2 = 59297
	ActiveStateStackatov2_2_3 = 60815
	EnterpriseDBPPAS9_1_2     = 9953
	EnterpriseDBPSQL9_1_3     = 9995
)

type SecurityGroup struct {
	Name string `json:"name"`
}

type Server struct {
	ConfigDrive    bool              `json:"config_drive"`
	FlavorRef      Flavor            `json:"flavorRef"`
	ImageRef       ServerImage       `json:"imageRef"`
	MaxCount       int               `json:"max_count"`
	MinCount       int               `json:"min_count"`
	Name           string            `json:"name"`
	Key            string            `json:"key_name"`
	Personality    string            `json:"personality"`
	UserData       string            `json:"user_data"`
	SecurityGroups []SecurityGroup   `json:"security_groups"`
	Metadata       map[string]string `json:"metadata"`
}

func (a Access) CreateServer(s Server) error {
	b, err := s.MarshalJSON()
	if err != nil {
		return err
	}
	path := fmt.Sprintf("%s%s/servers", COMPUTE_URL, a.TenantID)
	fmt.Println(path, string(b))
	client := &http.Client{}
	req, err := http.NewRequest("POST", path, strings.NewReader(string(b)))
	if err != nil {
		return err
	}
	req.Header.Add("X-Auth-Token", a.A.Token.ID)
	req.Header.Add("Content-type", "application/json")
	som, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return err
	}
	fmt.Println(string(som))
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(body))
	return nil
}

func (s Server) MarshalJSON() ([]byte, error) {
	b := bytes.NewBufferString("")
	b.WriteString(`{"server":{`)
	/* The available images are 100-105, x-small to x-large. */
	if s.FlavorRef < 100 || s.FlavorRef > 105 {
		return []byte{},
			errors.New("Flavor Reference refers to a non-existant flavour.")
	} else {
		b.WriteString(fmt.Sprintf(`"flavorRef":%d`, s.FlavorRef))
	}
	if s.ImageRef == 0 {
		return []byte{},
			errors.New("An image name is required.")
	} else {
		b.WriteString(fmt.Sprintf(`,"imageRef":%d`, s.ImageRef))
	}
	if s.Name == "" {
		return []byte{},
			errors.New("A name is required")
	} else {
		b.WriteString(fmt.Sprintf(`,"name":"%s"`, s.Name))
	}
	if s.Key != "" {
		b.WriteString(fmt.Sprintf(`,"key_name":"%s"`, s.Key))
	}
	if s.ConfigDrive {
		b.WriteString(`,"config_drive": true`)
	}
	if s.MinCount > 0 {
		b.WriteString(fmt.Sprintf(`,"min_count":%d`, s.MinCount))
	}
	if s.MaxCount > 0 {
		b.WriteString(fmt.Sprintf(`,"max_count":%d`, s.MaxCount))
	}
	if len(s.Personality) > 255 {
		return []byte{},
			errors.New("Server's personality cannot have >255 bytes.")
	} else if s.Personality != "" {
		b.WriteString(fmt.Sprintf(`,"personality":"%s",`, s.Personality))
	}
	if s.UserData != "" {
		newb := make([]byte, 0, len(s.UserData))
		base64.StdEncoding.Encode([]byte(s.UserData), newb)
		b.WriteString(fmt.Sprintf(`,"user_data": "%s",`, string(newb)))
	}
	if len(s.Metadata) > 0 {
		fmt.Println(len(s.Metadata))
		b.WriteString(`,"metadata":{`)
		cnt := 0
		for key, value := range s.Metadata {
			b.WriteString(fmt.Sprintf(`"%s": "%s"`, key, value))
			if cnt+1 != len(s.Metadata) {
				b.WriteString(",")
				cnt++
			} else {
				b.WriteString("}")
			}
		}
	}
	if len(s.SecurityGroups) > 0 {
		b.WriteString(`,"security_groups":[`)
		cnt := 0
		for _, sg := range s.SecurityGroups {
			b.WriteString(fmt.Sprintf(`{"name": "%s"}`, sg.Name))
			if cnt+1 != len(s.SecurityGroups) {
				b.WriteString(",")
				cnt++
			} else {
				b.WriteString("]")
			}
		}
	}
	b.WriteString("}}")
	return b.Bytes(), nil
}