package edgegrid

import (
	"fmt"
	"log"
	"strings"
	"time"
)

type DNSZone struct {
	service *ConfigDNSV1Service
	Token   string `json:"token"`
	Zone    struct {
		Name       string                  `json:"name,omitempty"`
		A          DNSRecordSet            `json:"a,omitempty"`
		AAAA       DNSRecordSet            `json:"aaaa,omitempty"`
		Afsdb      DNSRecordSet            `json:"afsdb,omitempty"`
		Cname      DNSRecordSet            `json:"cname,omitempty"`
		Dnskey     DNSRecordSet            `json:"dnskey,omitempty"`
		Ds         DNSRecordSet            `json:"ds,omitempty"`
		Hinfo      DNSRecordSet            `json:"hinfo,omitempty"`
		Loc        DNSRecordSet            `json:"loc,omitempty"`
		Mx         DNSRecordSet            `json:"mx,omitempty"`
		Naptr      DNSRecordSet            `json:"naptr,omitempty"`
		Ns         DNSRecordSet            `json:"ns,omitempty"`
		Nsec3      DNSRecordSet            `json:"nsec3,omitempty"`
		Nsec3param DNSRecordSet            `json:"nsec3param,omitempty"`
		Ptr        DNSRecordSet            `json:"ptr,omitempty"`
		Rp         DNSRecordSet            `json:"rp,omitempty"`
		Rrsig      DNSRecordSet            `json:"rrsig,omitempty"`
		Soa        *DNSRecord              `json:"soa,omitempty"`
		Spf        DNSRecordSet            `json:"spf,omitempty"`
		Srv        DNSRecordSet            `json:"srv,omitempty"`
		Sshfp      DNSRecordSet            `json:"sshfp,omitempty"`
		Txt        DNSRecordSet            `json:"txt,omitempty"`
		Records    map[string]DNSRecordSet `json:"-"`
	} `json:"zone"`
}

func (zone *DNSZone) Save() error {
	zone.unmarshalRecords()

	zone.Zone.Soa.Serial = int(time.Now().Unix())

	res, err := zone.service.client.PostJSON("/config-dns/v1/zones/"+zone.Zone.Name, zone)
	if err != nil {
		return err
	}

	if res.IsError() {
		err := NewAPIError(res)
		return fmt.Errorf("Unable to save record (%s)", err.Error())
	}

	for {
		updatedZone, err := zone.service.GetZone(zone.Zone.Name)
		if err != nil {
			return err
		}

		if updatedZone.Token != zone.Token {
			log.Printf("[TRACE] Token updated: old: %s, new: %s", zone.Token, updatedZone.Token)
			*zone = *updatedZone
			break
		}
		log.Println("[DEBUG] Token not updated, retrying...")
		time.Sleep(time.Second)
	}

	if err != nil {
		return fmt.Errorf("Unable to save record (%s)", err.Error())
	}

	log.Printf("[INFO] Zone Saved")

	return nil
}

func (zone *DNSZone) marshalRecords() {
	zone.Zone.Records = make(map[string]DNSRecordSet)
	zone.Zone.Records["A"] = zone.Zone.A
	zone.Zone.Records["AAAA"] = zone.Zone.AAAA
	zone.Zone.Records["AFSDB"] = zone.Zone.Afsdb
	zone.Zone.Records["CNAME"] = zone.Zone.Cname
	zone.Zone.Records["DNSKEY"] = zone.Zone.Dnskey
	zone.Zone.Records["DS"] = zone.Zone.Ds
	zone.Zone.Records["HINFO"] = zone.Zone.Hinfo
	zone.Zone.Records["LOC"] = zone.Zone.Loc
	zone.Zone.Records["MX"] = zone.Zone.Mx
	zone.Zone.Records["NAPTR"] = zone.Zone.Naptr
	zone.Zone.Records["NS"] = zone.Zone.Ns
	zone.Zone.Records["NSEC3"] = zone.Zone.Nsec3
	zone.Zone.Records["NSEC3PARAM"] = zone.Zone.Nsec3param
	zone.Zone.Records["PTR"] = zone.Zone.Ptr
	zone.Zone.Records["RP"] = zone.Zone.Rp
	zone.Zone.Records["RRSIG"] = zone.Zone.Rrsig
	zone.Zone.Records["SOA"] = []*DNSRecord{zone.Zone.Soa}
	zone.Zone.Records["SPF"] = zone.Zone.Spf
	zone.Zone.Records["SRV"] = zone.Zone.Srv
	zone.Zone.Records["SSHFP"] = zone.Zone.Sshfp
	zone.Zone.Records["TXT"] = zone.Zone.Txt
}

func (zone *DNSZone) unmarshalRecords() {
	zone.Zone.A = zone.Zone.Records["A"]
	zone.Zone.AAAA = zone.Zone.Records["AAAA"]
	zone.Zone.Afsdb = zone.Zone.Records["AFSDB"]
	zone.Zone.Cname = zone.Zone.Records["CNAME"]
	zone.Zone.Dnskey = zone.Zone.Records["DNSKEY"]
	zone.Zone.Ds = zone.Zone.Records["DS"]
	zone.Zone.Hinfo = zone.Zone.Records["HINFO"]
	zone.Zone.Loc = zone.Zone.Records["LOC"]
	zone.Zone.Mx = zone.Zone.Records["MX"]
	zone.Zone.Naptr = zone.Zone.Records["NAPTR"]
	zone.Zone.Ns = zone.Zone.Records["NS"]
	zone.Zone.Nsec3 = zone.Zone.Records["NSEC3"]
	zone.Zone.Nsec3param = zone.Zone.Records["NSEC3PARAM"]
	zone.Zone.Ptr = zone.Zone.Records["PTR"]
	zone.Zone.Rp = zone.Zone.Records["RP"]
	zone.Zone.Rrsig = zone.Zone.Records["RRSIG"]
	zone.Zone.Soa = zone.Zone.Records["SOA"][0]
	zone.Zone.Spf = zone.Zone.Records["SPF"]
	zone.Zone.Srv = zone.Zone.Records["SRV"]
	zone.Zone.Sshfp = zone.Zone.Records["SSHFP"]
	zone.Zone.Txt = zone.Zone.Records["TXT"]
}

func (zone *DNSZone) fixupCnames(record *DNSRecord) {
	if record.RecordType == "CNAME" {
		names := make(map[string]string, len(zone.Zone.Records["CNAME"]))
		for _, record := range zone.Zone.Records["CNAME"] {
			names[strings.ToUpper(record.Name)] = record.Name
		}

		for recordType, records := range zone.Zone.Records {
			if recordType == "CNAME" {
				continue
			}

			newRecords := DNSRecordSet{}
			for _, record := range records {
				if _, ok := names[record.Name]; !ok {
					newRecords = append(newRecords, record)
				} else {
					log.Printf(
						"[WARN] %s Record conflicts with CNAME \"%s\", %[1]s Record ignored.",
						recordType,
						names[strings.ToUpper(record.Name)],
					)
				}
			}
			zone.Zone.Records[recordType] = newRecords
		}
	} else if record.Name != "" {
		name := strings.ToLower(record.Name)

		newRecords := DNSRecordSet{}
		for _, cname := range zone.Zone.Records["CNAME"] {
			if strings.ToLower(cname.Name) != name {
				newRecords = append(newRecords, cname)
			} else {
				log.Printf(
					"[WARN] %s Record \"%s\" conflicts with existing CNAME \"%s\", removing CNAME",
					record.RecordType,
					record.Name,
					cname.Name,
				)
			}
		}

		zone.Zone.Records["CNAME"] = newRecords
	}
}
