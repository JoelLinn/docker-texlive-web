# texlive-web
Build monolithic latex files using simple HTTP requests.

To test, run this localy 
`docker run -it --rm -p 8080:8080 joellinn/texlive-web`
and navigate to
`http://localhost:8080/`


It however is meant to be used by other docker containers.
You simply post a *.tex file to this image at `http://hostname:8080/pdflatex` and it will send a response code 200 with the pdf in the body if it compiled sucessfully.
Either use post it to the body with `Content-Type: 'application/x-tex'` or as multipart form file field `tex`.


### Examples
#### Command line
```
curl -X POST --output ./test.pdf -F "tex=@test.tex" http://localhost:8080/pdflatex
curl -X POST --output ./test.pdf -H "Content-Type: application/x-tex"  --data-binary  @test.tex http://localhost:8080/pdflatex 
```
#### PHP
```
// tex document in $tex_doc
$tex_doc = ""

$ch = curl_init();
$options = array(
    CURLOPT_URL => "http://texlive-web-host:8080/pdflatex",
    CURLOPT_RETURNTRANSFER => TRUE,
    CURLOPT_POST => TRUE,
    CURLOPT_POSTFIELDS => $tex_doc,
    CURLOPT_HTTPHEADER => array('Content-Type: application/x-tex')
); // cURL options
curl_setopt_array($ch, $options);

// pdf as binary string:
$ifcpdf=curl_exec($ch);

if (curl_getinfo($ch, CURLINFO_HTTP_CODE) != 200) {
    // ERROR
}

curl_close ($ch);
```
