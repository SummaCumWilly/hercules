package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	//	"time"
	"flag"
	"github.com/lucacervasio/mosesacs/cwmp"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
	//"strings"
)

var num_cpes = flag.Int("n", 2, "how many CPEs should I emulate ?")

var AcsUrl = "http://localhost:9292/acs"

func runConnection(cpe cwmp.CPE) {
	//	fmt.Printf("[%s] connecting with state %s\n", cpe.SerialNumber, cpe.State)
	fmt.Printf("[%s] --> Starting connection to %s, sending Inform with eventCode %s\n", cpe.SerialNumber, AcsUrl, cpe.State)

	buf := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
               xmlns:soap-enc="http://schemas.xmlsoap.org/soap/encoding/" xmlns:cwmp="urn:dslforum-org:cwmp-1-0"
               xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
    <soap:Header/>
    <soap:Body soap:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
        <cwmp:Inform>
            <DeviceId>
                <Manufacturer>` + cpe.Manufacturer + `</Manufacturer>
                <OUI>` + cpe.OUI + `</OUI>
                <ProductClass>Router</ProductClass>
                <SerialNumber>` + cpe.SerialNumber + `</SerialNumber>
            </DeviceId>
            <Event>
                <EventStruct>
                    <EventCode>` + cpe.State + `</EventCode>
                    <CommandKey/>
                </EventStruct>
            </Event>
            <MaxEnvelopes>1</MaxEnvelopes>
            <CurrentTime>2003-01-01T05:36:55Z</CurrentTime>
            <RetryCount>0</RetryCount>
            <ParameterList soap-enc:arrayType="cwmp:ParameterValueStruct[7]">
                <ParameterValueStruct xsi:type="cwmp:ParameterValueStruct">
                    <Name>InternetGatewayDevice.DeviceInfo.HardwareVersion</Name>
                    <Value xsi:type="xsd:string">NGRG 2009</Value>
                </ParameterValueStruct>
                <ParameterValueStruct xsi:type="cwmp:ParameterValueStruct">
                    <Name>InternetGatewayDevice.DeviceInfo.ProvisioningCode</Name>
                    <Value xsi:type="xsd:string">ABCD</Value>
                </ParameterValueStruct>
                <ParameterValueStruct xsi:type="cwmp:ParameterValueStruct">
                    <Name>InternetGatewayDevice.DeviceInfo.SoftwareVersion</Name>
                    <Value xsi:type="xsd:string">` + cpe.SoftwareVersion + `</Value>
                </ParameterValueStruct>
                <ParameterValueStruct xsi:type="cwmp:ParameterValueStruct">
                    <Name>InternetGatewayDevice.DeviceInfo.SpecVersion</Name>
                    <Value xsi:type="xsd:string">1.0</Value>
                </ParameterValueStruct>
                <ParameterValueStruct xsi:type="cwmp:ParameterValueStruct">
                    <Name>InternetGatewayDevice.ManagementServer.ConnectionRequestURL</Name>
                    <Value xsi:type="xsd:string">http://10.19.0.` + cpe.SerialNumber + `:9600/` + cpe.SerialNumber + `</Value>
                </ParameterValueStruct>
                <ParameterValueStruct xsi:type="cwmp:ParameterValueStruct">
                    <Name>InternetGatewayDevice.ManagementServer.ParameterKey</Name>
                    <Value xsi:type="xsd:string"/>
                </ParameterValueStruct>
                <ParameterValueStruct xsi:type="cwmp:ParameterValueStruct">
                    <Name>InternetGatewayDevice.WANDevice.1.WANConnectionDevice.1.WANIPConnection.1.ExternalIPAddress
                    </Name>
                    <Value xsi:type="xsd:string">10.19.0.` + cpe.SerialNumber + `</Value>
                </ParameterValueStruct>
            </ParameterList>
        </cwmp:Inform>
    </soap:Body>
</soap:Envelope>`

	cookieJar, _ := cookiejar.New(nil)
	tr := &http.Transport{}
	client := &http.Client{Transport: tr, Jar: cookieJar}

	resp, err := client.Post(AcsUrl, "text/xml", bytes.NewBufferString(buf))
	if err != nil {
		fmt.Println(fmt.Sprintf("Couldn't connect to %s", AcsUrl))
		os.Exit(1)
	}

	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()

	fmt.Printf("[%s] <-- ACS replied with statusCode: %d, content-lenght: %d\n", cpe.SerialNumber, resp.StatusCode, resp.ContentLength)
	//	tmp,_ := ioutil.ReadAll(resp.Body)
	//	fmt.Printf("body: %s", string(tmp))

	fmt.Printf("[%s] --> Sending empty POST\n", cpe.SerialNumber)
	resp, err = client.Post(AcsUrl, "text/xml", bytes.NewBufferString(""))
	if err != nil {
		fmt.Println(err)
	}

	for {
		fmt.Printf("[%s] <-- ACS replied with statusCode: %d, content-lenght: %d\n", cpe.SerialNumber, resp.StatusCode, resp.ContentLength)
		if resp.StatusCode == 204 {
			break
		} else {
			// parse and reply
			tmp, _ := ioutil.ReadAll(resp.Body)
			body := string(tmp)

			//if strings.Contains(body, "GetParameterValues") {
			fmt.Println("Got GetParameterValues" + body)
			resp, err = client.Post(AcsUrl, "text/xml", bytes.NewBufferString(`<?xml version="1.0" encoding="UTF-8"?>
        <soap:Envelope xmlns:soapenc="http://schemas.xmlsoap.org/soap/encoding/" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:cwmp="urn:dslforum-org:cwmp-1-0" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:schemaLocation="urn:dslforum-org:cwmp-1-0 ..\schemas\wt121.xsd" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
        <soap:Header/>
        <soap:Body soap:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
          <cwmp:GetParameterValuesResponse>
              <ParameterrNames>
                    <string>Boh</string>
                        </ParameterNames>
                          </cwmp:GetParameterValuesResponse>
                          </soap:Body>
                          </soap:Envelope>`))
			//}
		}
	}

	resp.Body.Close()

	tr.CloseIdleConnections()
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("new connection Request")
}

func periodic(interval int, cpe cwmp.CPE) {
	fmt.Printf("Bootstrapping CPE #%s with interval %ds\n", cpe.SerialNumber, interval)
	runConnection(cpe)
	for {
		time.Sleep(time.Duration(interval) * time.Second)
		runConnection(cpe)
	}
}

func random(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

func main() {
	// create cpe struct

	flag.Parse()
	fmt.Println("Starting Hercules with", *num_cpes, "cpes")

	CPEs := []cwmp.CPE{}

	// initialize CPEs and send bootstrap
	//cpe1 := cwmp.CPE{"1", "PIRELLI BROADBAND SOLUTIONS", "0013C8", "asd", "asd", "asd", "0 BOOTSTRAP"}
	//	cpe2 := CPE{"2", "Telsey", "0014", "asd", "asd", "asd", "1 BOOT"}

	for i := 1; i <= *num_cpes; i++ {
		tmp_cpe := cwmp.CPE{strconv.Itoa(i), "PIRELLI BROADBAND SOLUTIONS", "0013C8", "asd", "asd", "asd", "0 BOOTSTRAP", nil}
		CPEs = append(CPEs, tmp_cpe)
	}

	//	fmt.Println(CPEs)

	for _, c := range CPEs {
		go periodic(random(10, 120), c)
	}

	// TODO run httpserver to wait for connection
	//	time.Sleep (3 * time.Second)
	http.HandleFunc("/acs", handler)
	fmt.Println("Listening connection request port on 9600")
	err := http.ListenAndServe(":9600", nil)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

}
