package rest

const(
documentation = `
{{ define "path-param" }}
    {{ $arr := Split ( js . ) "/" }}
    {{ js . }}
    <div class="row">
    {{ range $part := $arr }}
        {{ if eq $part "" }}
        {{ else if ( Contains $part ":" ) }}
            <div class="col-xs-2">
                <input id="{{ $part }}" class="form-control" type="text" placeholder="{{ $part }}">
            </div>
        {{ else }}
        <div class="col-xs-2">{{ $part }}</div>
        {{ end }}
    {{ end }}
    </div>
{{ end }}

{{ define "body-param" }}
    {{ if .HasBodyParam }}
    <div class="row">
        <div class="col-xs-6">
            <div class="form-group">
                <label for="body">Body</label>
                <textarea id="body" class="form-control" rows="10">{{ .JsonSchema }}
                </textarea>
            </div>
        </div>
    </div>
    {{ end }}
{{ end }}

{{$page := .}}

<div class="text-center">
    <h1> {{.Title}} REST API on {{$page.Host}}</h1>
</div>

<div class="container">
    {{range $route := .Routes}}
    <div>
        {{ if eq $route.Method "PUT" }}
            <div class="bg-warning">
                <form id="{{ printf "%s-%s" $route.Method $route.Path }}">
                    <input type="button"
                        class="btn btn-warning"
                        value="{{$route.Method}}"
                        onClick="doRequest('{{ $route.Method }}',
                            '{{ js (printf "%s-%s" $route.Method $route.Path) }}',
                            '{{ printf "http://%s%s" $page.Host $route.Path}}'
                    )">
                    {{ template "path-param" $route.Path }}
                    <br>
                    {{ template "body-param" $route }}
                </form>
            </div>
            <div id="{{$route.Method}}-{{$route.Path}}"></div>


        {{ else if eq $route.Method "POST" }}
            <div class="bg-success">
                <form id="{{ printf "%s-%s" $route.Method $route.Path }}">
                    <input type="button"
                        class="btn btn-success active"
                        value="{{$route.Method}}"
                        onClick="doRequest('{{ $route.Method }}',
                            '{{ js (printf "%s-%s" $route.Method $route.Path) }}',
                            '{{ printf "http://%s%s" $page.Host $route.Path}}'
                    )">
                    {{ template "path-param" $route.Path }}
                    <br>
                    {{ template "body-param" $route }}
                </form>
            </div>
            <div id="{{$route.Method}}-{{$route.Path}}"></div>



        {{ else if eq $route.Method "DELETE" }}
            <div class="bg-danger">
                <form id="{{ printf "%s-%s" $route.Method $route.Path }}">
                    <input type="button"
                        class="btn btn-danger active"
                        value="{{$route.Method}}"
                        onClick="doRequest('{{ $route.Method }}',
                            '{{ js (printf "%s-%s" $route.Method $route.Path) }}',
                            '{{ printf "http://%s%s" $page.Host $route.Path}}'
                    )">
                    {{ template "path-param" $route.Path }}
                    <br>
                    {{ template "body-param" $route }}
                </form>
            </div>
            <div id="{{$route.Method}}-{{$route.Path}}"></div>
            

        {{ else if eq $route.Method "GET"}}
            <div class="bg-info">
                <form id="{{ printf "%s-%s" $route.Method $route.Path }}">
                    <input type="button"
                        class="btn btn-info active"
                        value="{{$route.Method}}"
                        onClick="doRequest('{{ $route.Method }}',
                            '{{ js (printf "%s-%s" $route.Method $route.Path) }}',
                            '{{ printf "http://%s%s" $page.Host $route.Path}}'
                    )">
                    {{ template "path-param" $route.Path }}
                    <br>
                    {{ template "body-param" $route }}
                </form>
            </div>
            <div id="{{ printf "%s-%s" $route.Method $route.Path}}"></div>


        {{ else }}
            <p class="bg-primary">
                Unsupported HTTP Verb {{ $route.Method }}
            </p>
        {{ end }}
    </div>
    {{end}}
</div>

<script type="text/javascript">

    function showError (result, resultDiv) {
                console.log(result);
                resultDiv.html('<pre class="box">Status: ' + result.status
                    + "<br>Status Text: " + result.statusText
                    + "<br>Response: " + result.responseText + "</pre>");
    }

    function getJsDivID (divID) {
        divID = divID.replace("-", "\\-");
        divID = divID.replace(new RegExp(":", 'g'), "\\:");
        divID = divID.replace(new RegExp("/", 'g'), "\\/");
        return divID
    }

    function replaceInPath (divID, path, resultDiv) {
        resultDiv.html(""); // TODO: add a spinning wheel for in progress.

        formDiv = $("form#"+divID+" :input:text");

        for (var i = 0; i < formDiv.length; ++i) {
            path = path.replace(formDiv[i].id, formDiv[i].value);

            if (formDiv[i].value == "") {
                resultDiv.html('<pre class="box">text box for variable "' + formDiv[i].id + '" is empty</pre>');
                return;
            }
        }
        return path;
    }

    function getBody(divID, resultDiv) {
        textArea = $("form#"+divID+" textarea#body");
        console.log(textArea);
        if (textArea.length > 0) {
            return textArea[0].value
        } else {
            return
        }
    }

    function doRequest(verb, divID, path) {
        divID = getJsDivID(divID);
        resultDiv = $("div#" + divID);
        path = replaceInPath(divID, path, resultDiv);
        if (!path) { return; }

        var request = {
            url: path,
            type: verb,
            success: function(result, textStatus, jqXHR) {
                console.log(jqXHR);
                var js = "Status: " + jqXHR.status + "\n";
                js += "Status Text: " + jqXHR.statusText;
                if (result) {
                    js += "\n\n" + JSON.stringify(result, null, 2);
                }
                resultDiv.html("<pre class=box>" + js + "</pre>");
            },
            error: function(jqXHR) {
                showError(jqXHR, resultDiv);
            }
        }

        if (verb == 'POST' || verb == 'PUT') {
            var body = getBody(divID, resultDiv);
            request.headers = { "Content-Type": "application/json" };
            request.data = body;
        } else if (verb == 'DELETE') {
            request.headers = { "Content-Type": "application/json" };
        }

        $.ajax(request);
    }

</script>


<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/css/bootstrap.min.css">
<script src="https://code.jquery.com/jquery-2.1.4.min.js"></script>
<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/js/bootstrap.min.js"></script>
`
)
