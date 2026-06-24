// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package coredata

import (
	"database/sql/driver"
	"encoding"
	"fmt"
	"strings"
)

type (
	CountryCode string
)

const (
	CountryCodeGlobal CountryCode = "GLOBAL"
)

const (
	CountryCodeAD CountryCode = "AD"
	CountryCodeAE CountryCode = "AE"
	CountryCodeAF CountryCode = "AF"
	CountryCodeAG CountryCode = "AG"
	CountryCodeAI CountryCode = "AI"
	CountryCodeAL CountryCode = "AL"
	CountryCodeAM CountryCode = "AM"
	CountryCodeAO CountryCode = "AO"
	CountryCodeAQ CountryCode = "AQ"
	CountryCodeAR CountryCode = "AR"
	CountryCodeAS CountryCode = "AS"
	CountryCodeAT CountryCode = "AT"
	CountryCodeAU CountryCode = "AU"
	CountryCodeAW CountryCode = "AW"
	CountryCodeAX CountryCode = "AX"
	CountryCodeAZ CountryCode = "AZ"
	CountryCodeBA CountryCode = "BA"
	CountryCodeBB CountryCode = "BB"
	CountryCodeBD CountryCode = "BD"
	CountryCodeBE CountryCode = "BE"
	CountryCodeBF CountryCode = "BF"
	CountryCodeBG CountryCode = "BG"
	CountryCodeBH CountryCode = "BH"
	CountryCodeBI CountryCode = "BI"
	CountryCodeBJ CountryCode = "BJ"
	CountryCodeBL CountryCode = "BL"
	CountryCodeBM CountryCode = "BM"
	CountryCodeBN CountryCode = "BN"
	CountryCodeBO CountryCode = "BO"
	CountryCodeBQ CountryCode = "BQ"
	CountryCodeBR CountryCode = "BR"
	CountryCodeBS CountryCode = "BS"
	CountryCodeBT CountryCode = "BT"
	CountryCodeBV CountryCode = "BV"
	CountryCodeBW CountryCode = "BW"
	CountryCodeBY CountryCode = "BY"
	CountryCodeBZ CountryCode = "BZ"
	CountryCodeCA CountryCode = "CA"
	CountryCodeCC CountryCode = "CC"
	CountryCodeCD CountryCode = "CD"
	CountryCodeCF CountryCode = "CF"
	CountryCodeCG CountryCode = "CG"
	CountryCodeCH CountryCode = "CH"
	CountryCodeCI CountryCode = "CI"
	CountryCodeCK CountryCode = "CK"
	CountryCodeCL CountryCode = "CL"
	CountryCodeCM CountryCode = "CM"
	CountryCodeCN CountryCode = "CN"
	CountryCodeCO CountryCode = "CO"
	CountryCodeCR CountryCode = "CR"
	CountryCodeCU CountryCode = "CU"
	CountryCodeCV CountryCode = "CV"
	CountryCodeCW CountryCode = "CW"
	CountryCodeCX CountryCode = "CX"
	CountryCodeCY CountryCode = "CY"
	CountryCodeCZ CountryCode = "CZ"
	CountryCodeDE CountryCode = "DE"
	CountryCodeDJ CountryCode = "DJ"
	CountryCodeDK CountryCode = "DK"
	CountryCodeDM CountryCode = "DM"
	CountryCodeDO CountryCode = "DO"
	CountryCodeDZ CountryCode = "DZ"
	CountryCodeEC CountryCode = "EC"
	CountryCodeEE CountryCode = "EE"
	CountryCodeEG CountryCode = "EG"
	CountryCodeEH CountryCode = "EH"
	CountryCodeER CountryCode = "ER"
	CountryCodeES CountryCode = "ES"
	CountryCodeET CountryCode = "ET"
	CountryCodeEU CountryCode = "EU"
	CountryCodeFI CountryCode = "FI"
	CountryCodeFJ CountryCode = "FJ"
	CountryCodeFK CountryCode = "FK"
	CountryCodeFM CountryCode = "FM"
	CountryCodeFO CountryCode = "FO"
	CountryCodeFR CountryCode = "FR"
	CountryCodeGA CountryCode = "GA"
	CountryCodeGB CountryCode = "GB"
	CountryCodeGD CountryCode = "GD"
	CountryCodeGE CountryCode = "GE"
	CountryCodeGF CountryCode = "GF"
	CountryCodeGG CountryCode = "GG"
	CountryCodeGH CountryCode = "GH"
	CountryCodeGI CountryCode = "GI"
	CountryCodeGL CountryCode = "GL"
	CountryCodeGM CountryCode = "GM"
	CountryCodeGN CountryCode = "GN"
	CountryCodeGP CountryCode = "GP"
	CountryCodeGQ CountryCode = "GQ"
	CountryCodeGR CountryCode = "GR"
	CountryCodeGT CountryCode = "GT"
	CountryCodeGU CountryCode = "GU"
	CountryCodeGW CountryCode = "GW"
	CountryCodeGY CountryCode = "GY"
	CountryCodeHK CountryCode = "HK"
	CountryCodeHM CountryCode = "HM"
	CountryCodeHN CountryCode = "HN"
	CountryCodeHR CountryCode = "HR"
	CountryCodeHT CountryCode = "HT"
	CountryCodeHU CountryCode = "HU"
	CountryCodeID CountryCode = "ID"
	CountryCodeIE CountryCode = "IE"
	CountryCodeIL CountryCode = "IL"
	CountryCodeIM CountryCode = "IM"
	CountryCodeIN CountryCode = "IN"
	CountryCodeIO CountryCode = "IO"
	CountryCodeIQ CountryCode = "IQ"
	CountryCodeIR CountryCode = "IR"
	CountryCodeIS CountryCode = "IS"
	CountryCodeIT CountryCode = "IT"
	CountryCodeJE CountryCode = "JE"
	CountryCodeJM CountryCode = "JM"
	CountryCodeJO CountryCode = "JO"
	CountryCodeJP CountryCode = "JP"
	CountryCodeKE CountryCode = "KE"
	CountryCodeKG CountryCode = "KG"
	CountryCodeKH CountryCode = "KH"
	CountryCodeKI CountryCode = "KI"
	CountryCodeKM CountryCode = "KM"
	CountryCodeKN CountryCode = "KN"
	CountryCodeKP CountryCode = "KP"
	CountryCodeKR CountryCode = "KR"
	CountryCodeKW CountryCode = "KW"
	CountryCodeKY CountryCode = "KY"
	CountryCodeKZ CountryCode = "KZ"
	CountryCodeLA CountryCode = "LA"
	CountryCodeLB CountryCode = "LB"
	CountryCodeLC CountryCode = "LC"
	CountryCodeLI CountryCode = "LI"
	CountryCodeLK CountryCode = "LK"
	CountryCodeLR CountryCode = "LR"
	CountryCodeLS CountryCode = "LS"
	CountryCodeLT CountryCode = "LT"
	CountryCodeLU CountryCode = "LU"
	CountryCodeLV CountryCode = "LV"
	CountryCodeLY CountryCode = "LY"
	CountryCodeMA CountryCode = "MA"
	CountryCodeMC CountryCode = "MC"
	CountryCodeMD CountryCode = "MD"
	CountryCodeME CountryCode = "ME"
	CountryCodeMF CountryCode = "MF"
	CountryCodeMG CountryCode = "MG"
	CountryCodeMH CountryCode = "MH"
	CountryCodeMK CountryCode = "MK"
	CountryCodeML CountryCode = "ML"
	CountryCodeMM CountryCode = "MM"
	CountryCodeMN CountryCode = "MN"
	CountryCodeMO CountryCode = "MO"
	CountryCodeMP CountryCode = "MP"
	CountryCodeMQ CountryCode = "MQ"
	CountryCodeMR CountryCode = "MR"
	CountryCodeMS CountryCode = "MS"
	CountryCodeMT CountryCode = "MT"
	CountryCodeMU CountryCode = "MU"
	CountryCodeMV CountryCode = "MV"
	CountryCodeMW CountryCode = "MW"
	CountryCodeMX CountryCode = "MX"
	CountryCodeMY CountryCode = "MY"
	CountryCodeMZ CountryCode = "MZ"
	CountryCodeNA CountryCode = "NA"
	CountryCodeNC CountryCode = "NC"
	CountryCodeNE CountryCode = "NE"
	CountryCodeNF CountryCode = "NF"
	CountryCodeNG CountryCode = "NG"
	CountryCodeNI CountryCode = "NI"
	CountryCodeNL CountryCode = "NL"
	CountryCodeNO CountryCode = "NO"
	CountryCodeNP CountryCode = "NP"
	CountryCodeNR CountryCode = "NR"
	CountryCodeNU CountryCode = "NU"
	CountryCodeNZ CountryCode = "NZ"
	CountryCodeOM CountryCode = "OM"
	CountryCodePA CountryCode = "PA"
	CountryCodePE CountryCode = "PE"
	CountryCodePF CountryCode = "PF"
	CountryCodePG CountryCode = "PG"
	CountryCodePH CountryCode = "PH"
	CountryCodePK CountryCode = "PK"
	CountryCodePL CountryCode = "PL"
	CountryCodePM CountryCode = "PM"
	CountryCodePN CountryCode = "PN"
	CountryCodePR CountryCode = "PR"
	CountryCodePS CountryCode = "PS"
	CountryCodePT CountryCode = "PT"
	CountryCodePW CountryCode = "PW"
	CountryCodePY CountryCode = "PY"
	CountryCodeQA CountryCode = "QA"
	CountryCodeRE CountryCode = "RE"
	CountryCodeRO CountryCode = "RO"
	CountryCodeRS CountryCode = "RS"
	CountryCodeRU CountryCode = "RU"
	CountryCodeRW CountryCode = "RW"
	CountryCodeSA CountryCode = "SA"
	CountryCodeSB CountryCode = "SB"
	CountryCodeSC CountryCode = "SC"
	CountryCodeSD CountryCode = "SD"
	CountryCodeSE CountryCode = "SE"
	CountryCodeSG CountryCode = "SG"
	CountryCodeSH CountryCode = "SH"
	CountryCodeSI CountryCode = "SI"
	CountryCodeSJ CountryCode = "SJ"
	CountryCodeSK CountryCode = "SK"
	CountryCodeSL CountryCode = "SL"
	CountryCodeSM CountryCode = "SM"
	CountryCodeSN CountryCode = "SN"
	CountryCodeSO CountryCode = "SO"
	CountryCodeSR CountryCode = "SR"
	CountryCodeSS CountryCode = "SS"
	CountryCodeST CountryCode = "ST"
	CountryCodeSV CountryCode = "SV"
	CountryCodeSX CountryCode = "SX"
	CountryCodeSY CountryCode = "SY"
	CountryCodeSZ CountryCode = "SZ"
	CountryCodeTC CountryCode = "TC"
	CountryCodeTD CountryCode = "TD"
	CountryCodeTF CountryCode = "TF"
	CountryCodeTG CountryCode = "TG"
	CountryCodeTH CountryCode = "TH"
	CountryCodeTJ CountryCode = "TJ"
	CountryCodeTK CountryCode = "TK"
	CountryCodeTL CountryCode = "TL"
	CountryCodeTM CountryCode = "TM"
	CountryCodeTN CountryCode = "TN"
	CountryCodeTO CountryCode = "TO"
	CountryCodeTR CountryCode = "TR"
	CountryCodeTT CountryCode = "TT"
	CountryCodeTV CountryCode = "TV"
	CountryCodeTW CountryCode = "TW"
	CountryCodeTZ CountryCode = "TZ"
	CountryCodeUA CountryCode = "UA"
	CountryCodeUG CountryCode = "UG"
	CountryCodeUM CountryCode = "UM"
	CountryCodeUS CountryCode = "US"
	CountryCodeUY CountryCode = "UY"
	CountryCodeUZ CountryCode = "UZ"
	CountryCodeVA CountryCode = "VA"
	CountryCodeVC CountryCode = "VC"
	CountryCodeVE CountryCode = "VE"
	CountryCodeVG CountryCode = "VG"
	CountryCodeVI CountryCode = "VI"
	CountryCodeVN CountryCode = "VN"
	CountryCodeVU CountryCode = "VU"
	CountryCodeWF CountryCode = "WF"
	CountryCodeWS CountryCode = "WS"
	CountryCodeYE CountryCode = "YE"
	CountryCodeYT CountryCode = "YT"
	CountryCodeZA CountryCode = "ZA"
	CountryCodeZM CountryCode = "ZM"
	CountryCodeZW CountryCode = "ZW"
)

var (
	_ fmt.Stringer             = CountryCode("")
	_ encoding.TextMarshaler   = CountryCode("")
	_ encoding.TextUnmarshaler = (*CountryCode)(nil)
)

func (v CountryCode) IsValid() bool {
	switch v {
	case
		CountryCodeGlobal,
		CountryCodeAD,
		CountryCodeAE,
		CountryCodeAF,
		CountryCodeAG,
		CountryCodeAI,
		CountryCodeAL,
		CountryCodeAM,
		CountryCodeAO,
		CountryCodeAQ,
		CountryCodeAR,
		CountryCodeAS,
		CountryCodeAT,
		CountryCodeAU,
		CountryCodeAW,
		CountryCodeAX,
		CountryCodeAZ,
		CountryCodeBA,
		CountryCodeBB,
		CountryCodeBD,
		CountryCodeBE,
		CountryCodeBF,
		CountryCodeBG,
		CountryCodeBH,
		CountryCodeBI,
		CountryCodeBJ,
		CountryCodeBL,
		CountryCodeBM,
		CountryCodeBN,
		CountryCodeBO,
		CountryCodeBQ,
		CountryCodeBR,
		CountryCodeBS,
		CountryCodeBT,
		CountryCodeBV,
		CountryCodeBW,
		CountryCodeBY,
		CountryCodeBZ,
		CountryCodeCA,
		CountryCodeCC,
		CountryCodeCD,
		CountryCodeCF,
		CountryCodeCG,
		CountryCodeCH,
		CountryCodeCI,
		CountryCodeCK,
		CountryCodeCL,
		CountryCodeCM,
		CountryCodeCN,
		CountryCodeCO,
		CountryCodeCR,
		CountryCodeCU,
		CountryCodeCV,
		CountryCodeCW,
		CountryCodeCX,
		CountryCodeCY,
		CountryCodeCZ,
		CountryCodeDE,
		CountryCodeDJ,
		CountryCodeDK,
		CountryCodeDM,
		CountryCodeDO,
		CountryCodeDZ,
		CountryCodeEC,
		CountryCodeEE,
		CountryCodeEG,
		CountryCodeEH,
		CountryCodeER,
		CountryCodeES,
		CountryCodeET,
		CountryCodeEU,
		CountryCodeFI,
		CountryCodeFJ,
		CountryCodeFK,
		CountryCodeFM,
		CountryCodeFO,
		CountryCodeFR,
		CountryCodeGA,
		CountryCodeGB,
		CountryCodeGD,
		CountryCodeGE,
		CountryCodeGF,
		CountryCodeGG,
		CountryCodeGH,
		CountryCodeGI,
		CountryCodeGL,
		CountryCodeGM,
		CountryCodeGN,
		CountryCodeGP,
		CountryCodeGQ,
		CountryCodeGR,
		CountryCodeGT,
		CountryCodeGU,
		CountryCodeGW,
		CountryCodeGY,
		CountryCodeHK,
		CountryCodeHM,
		CountryCodeHN,
		CountryCodeHR,
		CountryCodeHT,
		CountryCodeHU,
		CountryCodeID,
		CountryCodeIE,
		CountryCodeIL,
		CountryCodeIM,
		CountryCodeIN,
		CountryCodeIO,
		CountryCodeIQ,
		CountryCodeIR,
		CountryCodeIS,
		CountryCodeIT,
		CountryCodeJE,
		CountryCodeJM,
		CountryCodeJO,
		CountryCodeJP,
		CountryCodeKE,
		CountryCodeKG,
		CountryCodeKH,
		CountryCodeKI,
		CountryCodeKM,
		CountryCodeKN,
		CountryCodeKP,
		CountryCodeKR,
		CountryCodeKW,
		CountryCodeKY,
		CountryCodeKZ,
		CountryCodeLA,
		CountryCodeLB,
		CountryCodeLC,
		CountryCodeLI,
		CountryCodeLK,
		CountryCodeLR,
		CountryCodeLS,
		CountryCodeLT,
		CountryCodeLU,
		CountryCodeLV,
		CountryCodeLY,
		CountryCodeMA,
		CountryCodeMC,
		CountryCodeMD,
		CountryCodeME,
		CountryCodeMF,
		CountryCodeMG,
		CountryCodeMH,
		CountryCodeMK,
		CountryCodeML,
		CountryCodeMM,
		CountryCodeMN,
		CountryCodeMO,
		CountryCodeMP,
		CountryCodeMQ,
		CountryCodeMR,
		CountryCodeMS,
		CountryCodeMT,
		CountryCodeMU,
		CountryCodeMV,
		CountryCodeMW,
		CountryCodeMX,
		CountryCodeMY,
		CountryCodeMZ,
		CountryCodeNA,
		CountryCodeNC,
		CountryCodeNE,
		CountryCodeNF,
		CountryCodeNG,
		CountryCodeNI,
		CountryCodeNL,
		CountryCodeNO,
		CountryCodeNP,
		CountryCodeNR,
		CountryCodeNU,
		CountryCodeNZ,
		CountryCodeOM,
		CountryCodePA,
		CountryCodePE,
		CountryCodePF,
		CountryCodePG,
		CountryCodePH,
		CountryCodePK,
		CountryCodePL,
		CountryCodePM,
		CountryCodePN,
		CountryCodePR,
		CountryCodePS,
		CountryCodePT,
		CountryCodePW,
		CountryCodePY,
		CountryCodeQA,
		CountryCodeRE,
		CountryCodeRO,
		CountryCodeRS,
		CountryCodeRU,
		CountryCodeRW,
		CountryCodeSA,
		CountryCodeSB,
		CountryCodeSC,
		CountryCodeSD,
		CountryCodeSE,
		CountryCodeSG,
		CountryCodeSH,
		CountryCodeSI,
		CountryCodeSJ,
		CountryCodeSK,
		CountryCodeSL,
		CountryCodeSM,
		CountryCodeSN,
		CountryCodeSO,
		CountryCodeSR,
		CountryCodeSS,
		CountryCodeST,
		CountryCodeSV,
		CountryCodeSX,
		CountryCodeSY,
		CountryCodeSZ,
		CountryCodeTC,
		CountryCodeTD,
		CountryCodeTF,
		CountryCodeTG,
		CountryCodeTH,
		CountryCodeTJ,
		CountryCodeTK,
		CountryCodeTL,
		CountryCodeTM,
		CountryCodeTN,
		CountryCodeTO,
		CountryCodeTR,
		CountryCodeTT,
		CountryCodeTV,
		CountryCodeTW,
		CountryCodeTZ,
		CountryCodeUA,
		CountryCodeUG,
		CountryCodeUM,
		CountryCodeUS,
		CountryCodeUY,
		CountryCodeUZ,
		CountryCodeVA,
		CountryCodeVC,
		CountryCodeVE,
		CountryCodeVG,
		CountryCodeVI,
		CountryCodeVN,
		CountryCodeVU,
		CountryCodeWF,
		CountryCodeWS,
		CountryCodeYE,
		CountryCodeYT,
		CountryCodeZA,
		CountryCodeZM,
		CountryCodeZW:
		return true
	}

	return false
}

func (v CountryCode) String() string {
	return string(v)
}

func (v CountryCode) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CountryCode) UnmarshalText(text []byte) error {
	val := CountryCode(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CountryCode value: %q", string(text))
	}

	*v = val

	return nil
}

type CountryCodes []CountryCode

func (s *CountryCodes) Scan(value any) error {
	switch v := value.(type) {
	case string:
		return s.scanFromString(v)
	case []byte:
		return s.scanFromString(string(v))
	default:
		return fmt.Errorf("unsupported type for CountryCodes: %T", value)
	}
}

func (s *CountryCodes) scanFromString(str string) error {
	str = strings.TrimSpace(str)
	if str == "{}" || str == "" {
		*s = []CountryCode{}
		return nil
	}

	if strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}") {
		str = str[1 : len(str)-1]
	}

	parts := strings.Split(str, ",")
	result := make([]CountryCode, len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part)

		if strings.HasPrefix(part, `"`) && strings.HasSuffix(part, `"`) {
			part = part[1 : len(part)-1]
		}

		var ct CountryCode
		if err := ct.UnmarshalText([]byte(part)); err != nil {
			return fmt.Errorf("invalid country code in array: %s", part)
		}

		result[i] = ct
	}

	*s = result

	return nil
}

func (s CountryCodes) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "{}", nil
	}

	values := make([]string, len(s))
	for i, ct := range s {
		values[i] = ct.String()
	}

	return "{" + strings.Join(values, ",") + "}", nil
}
