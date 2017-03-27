package edgegrid

import (
	"errors"
	"log"
	"strings"
	"time"
)

type DnsZone struct {
	service *ConfigDnsV1Service
	Token   string `json:"token"`
	Zone    struct {
		Name       string                  `json:"name,omitempty"`
		A          DnsRecordSet            `json:"a,omitempty"`
		AAAA       DnsRecordSet            `json:"aaaa,omitempty"`
		Afsdb      DnsRecordSet            `json:"afsdb,omitempty"`
		Cname      DnsRecordSet            `json:"cname,omitempty"`
		Dnskey     DnsRecordSet            `json:"dnskey,omitempty"`
		Ds         DnsRecordSet            `json:"ds,omitempty"`
		Hinfo      DnsRecordSet            `json:"hinfo,omitempty"`
		Loc        DnsRecordSet            `json:"loc,omitempty"`
		Mx         DnsRecordSet            `json:"mx,omitempty"`
		Naptr      DnsRecordSet            `json:"naptr,omitempty"`
		Ns         DnsRecordSet            `json:"ns,omitempty"`
		Nsec3      DnsRecordSet            `json:"nsec3,omitempty"`
		Nsec3param DnsRecordSet            `json:"nsec3param,omitempty"`
		Ptr        DnsRecordSet            `json:"ptr,omitempty"`
		Rp         DnsRecordSet            `json:"rp,omitempty"`
		Rrsig      DnsRecordSet            `json:"rrsig,omitempty"`
		Soa        *DnsRecord              `json:"soa,omitempty"`
		Spf        DnsRecordSet            `json:"spf,omitempty"`
		Srv        DnsRecordSet            `json:"srv,omitempty"`
		Sshfp      DnsRecordSet            `json:"sshfp,omitempty"`
		Txt        DnsRecordSet            `json:"txt,omitempty"`
		Records    map[string]DnsRecordSet `json:"-"`
	} `json:"zone"`
}

func (zone *DnsZone) Save() error {
	zone.unmarshalRecords()

	zone.Zone.Soa.Serial = int(time.Now().Unix())

	res, err := zone.service.client.PostJson("/config-dns/v1/zones/"+zone.Zone.Name, zone)
	if err != nil {
		return err
	}

	if res.IsError() == true {
		err := NewApiError(res)
		return errors.New("Unable to save record (" + err.Error() + ")")
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
		return errors.New("Unable to save record (" + err.Error() + ")")
	}

	log.Printf("[INFO] Zone Saved")

	return nil
}

func (zone *DnsZone) marshalRecords() {
	zone.Zone.Records = make(map[string]DnsRecordSet)
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
	zone.Zone.Records["SOA"] = []*DnsRecord{zone.Zone.Soa}
	zone.Zone.Records["SPF"] = zone.Zone.Spf
	zone.Zone.Records["SRV"] = zone.Zone.Srv
	zone.Zone.Records["SSHFP"] = zone.Zone.Sshfp
	zone.Zone.Records["TXT"] = zone.Zone.Txt
}

func (zone *DnsZone) unmarshalRecords() {
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

func (zone *DnsZone) fixupCnames(record *DnsRecord) {
	if record.RecordType == "CNAME" {
		names := make(map[string]string, len(zone.Zone.Records["CNAME"]))
		for _, record := range zone.Zone.Records["CNAME"] {
			names[strings.ToUpper(record.Name)] = record.Name
		}

		for recordType, records := range zone.Zone.Records {
			if recordType == "CNAME" {
				continue
			}

			newRecords := DnsRecordSet{}
			for _, record := range records {
				if _, ok := names[record.Name]; ok == false {
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

		newRecords := DnsRecordSet{}
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
