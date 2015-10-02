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
                        onClick="doPut(this.form,
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
                        onClick="doPost(this.form,
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
                        onClick="doDelete(this.form,
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
                        onClick="doGet(this.form,
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
                resultDiv.html("Status: " + result.status
                    + "<br>Status Text: " + result.statusText
                    + "<br>Response: " + result.responseText);
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
                resultDiv.html("text box for variable '" + formDiv[i].id + "' is empty");
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

    function doGet(divID, path) {
        divID = getJsDivID(divID);
        resultDiv = $("div#"+divID);
        console.log(resultDiv);
        path = replaceInPath(divID, path, resultDiv);
        console.log(path);
        if (!path) { return; }

        $.ajax({
            url: path,
            type: 'GET',
            success: function (result) {
                console.log(result);
                resultDiv.html(result);
            },
            error: function(jqXHR) {
                showError(jqXHR, resultDiv);
            }
        });
    }

    function doPut(form, divID, path) {
        divID = getJsDivID(divID);
        resultDiv = $("div#"+divID);
        path = replaceInPath(divID, path, resultDiv);
        if (!path) { return; }
        
        $.ajax({
            url: path,
            type: 'PUT',
            headers: { "Content-Type": "application/json" }, 
            success: function(result) {
                resultDiv.html(result);
            },
            error: function(jqXHR) {
                showError(jqXHR, resultDiv);
            }
        });
    }

    function doPost(form, divID, path) {
        divID = getJsDivID(divID);
        resultDiv = $("div#"+divID);
        path = replaceInPath(divID, path, resultDiv);
        if (!path) { return; }

        var bod = getBody(divID, resultDiv);
        console.log(bod)
        
        $.ajax({
            url: path,
            type: 'POST',
            data: bod,
            headers: { "Content-Type": "application/json" }, 
            success: function(result) {
                resultDiv.html(result);
            },
            error: function(jqXHR) {
                showError(jqXHR, resultDiv);
            }
        });
    }

    function doDelete(form, divID, path) {
        divID = getJsDivID(divID);
        resultDiv = $("div#"+divID);
        path = replaceInPath(divID, path, resultDiv);
        if (!path) { return; }

        $.ajax({
            url: path,
            type: 'DELETE',
            headers: { "Content-Type": "application/json" }, 
            success: function(result) {
                resultDiv.html(result);
            },
            error: function(jqXHR) {
                showError(jqXHR, resultDiv);
            }
        });
    }
</script>


<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/css/bootstrap.min.css">
<script src="https://code.jquery.com/jquery-2.1.4.min.js"></script>
<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/js/bootstrap.min.js"></script>
`
)
