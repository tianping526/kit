package add

const protoTemplate = `
{{- /* delete empty line */ -}}
syntax = "proto3";

package {{.Package}};

{{ if eq .ServiceType "Interface" -}}
import "google/api/annotations.proto";
import "validate/validate.proto";
{{- else if eq .ServiceType "Admin" -}}
import "google/api/annotations.proto";
import "validate/validate.proto";
{{- else -}}
import "validate/validate.proto";
{{- end }}

option go_package = "{{.GoPackage}}";
option java_multiple_files = true;
option java_package = "{{.JavaPackage}}";

//
service {{.Service}}{{.ServiceType}} {
  //
{{- if eq .ServiceType "Interface" }}
  rpc Create{{.Service}}(Create{{.Service}}Request) returns (Create{{.Service}}Reply) {
    option (google.api.http) = {
      post: "{{.HTTPPath}}"
      body: "*"
    };
  }
{{- else if eq .ServiceType "Admin" }}
  rpc Create{{.Service}}(Create{{.Service}}Request) returns (Create{{.Service}}Reply) {
    option (google.api.http) = {
      post: "{{.HTTPPath}}"
      body: "*"
    };
  }
{{- else }}
  rpc Create{{.Service}}(Create{{.Service}}Request) returns (Create{{.Service}}Reply);
{{- end }}

  //
{{- if eq .ServiceType "Interface" }}
  rpc Update{{.Service}}(Update{{.Service}}Request) returns (Update{{.Service}}Reply) {
    option (google.api.http) = {
      patch: "{{.HTTPPath}}"
      body: "*"
    };
  }
{{- else if eq .ServiceType "Admin" }}
  rpc Update{{.Service}}(Update{{.Service}}Request) returns (Update{{.Service}}Reply) {
    option (google.api.http) = {
      patch: "{{.HTTPPath}}"
      body: "*"
    };
  }
{{- else }}
  rpc Update{{.Service}}(Update{{.Service}}Request) returns (Update{{.Service}}Reply);
{{- end }}

  //
{{- if eq .ServiceType "Interface" }}
  rpc Delete{{.Service}}(Delete{{.Service}}Request) returns (Delete{{.Service}}Reply) {
    option (google.api.http) = {delete: "{{.HTTPPath}}"};
  }
{{- else if eq .ServiceType "Admin" }}
  rpc Delete{{.Service}}(Delete{{.Service}}Request) returns (Delete{{.Service}}Reply) {
    option (google.api.http) = {delete: "{{.HTTPPath}}"};
  }
{{- else }}
  rpc Delete{{.Service}}(Delete{{.Service}}Request) returns (Delete{{.Service}}Reply);
{{- end }}

  //
{{- if eq .ServiceType "Interface" }}
  rpc Get{{.Service}}(Get{{.Service}}Request) returns (Get{{.Service}}Reply) {
    option (google.api.http) = {get: "{{.HTTPPath}}"};
  }
{{- else if eq .ServiceType "Admin" }}
  rpc Get{{.Service}}(Get{{.Service}}Request) returns (Get{{.Service}}Reply) {
    option (google.api.http) = {get: "{{.HTTPPath}}"};
  }
{{- else }}
  rpc Get{{.Service}}(Get{{.Service}}Request) returns (Get{{.Service}}Reply);
{{- end }}

  //
{{- if eq .ServiceType "Interface" }}
  rpc List{{.Service}}(List{{.Service}}Request) returns (List{{.Service}}Reply) {
    option (google.api.http) = {get: "{{.HTTPPath}}s"};
  }
{{- else if eq .ServiceType "Admin" }}
  rpc List{{.Service}}(List{{.Service}}Request) returns (List{{.Service}}Reply) {
    option (google.api.http) = {get: "{{.HTTPPath}}s"};
  }
{{- else }}
  rpc List{{.Service}}(List{{.Service}}Request) returns (List{{.Service}}Reply);
{{- end }}
}

//
message Create{{.Service}}Request {
  // xxx, max 64 bytes
  string xxx = 1 [(validate.rules).string = {max_bytes: 64}];
}

//
message Create{{.Service}}Reply {}

//
message Update{{.Service}}Request {
  // xxx, max 64 bytes
  string xxx = 1 [(validate.rules).string = {max_bytes: 64}];
}

//
message Update{{.Service}}Reply {}

//
message Delete{{.Service}}Request {
  // xxx, max 64 bytes
  string xxx = 1 [(validate.rules).string = {max_bytes: 64}];
}

//
message Delete{{.Service}}Reply {}

//
message Get{{.Service}}Request {
  // xxx, max 64 bytes
  string xxx = 1 [(validate.rules).string = {max_bytes: 64}];
}

//
message Get{{.Service}}Reply {}

//
message List{{.Service}}Request {
  // xxx, max 64 bytes
  string xxx = 1 [(validate.rules).string = {max_bytes: 64}];
}

//
message List{{.Service}}Reply {}
`

const errorProtoTemplate = `
{{- /* delete empty line */ -}}
syntax = "proto3";

package {{.Package}};

import "errors/errors.proto";

option go_package = "{{.GoPackage}}";
option java_multiple_files = true;
option java_package = "{{.JavaPackage}}";

// enumerate all error reasons for {{.Service}}{{.ServiceType}} service
enum {{.Service}}{{.ServiceType}}ErrorReason {
  option (errors.default_code) = 500;

  // unspecified value, not used, service internal error Reason field is generally an empty string
  ERROR_REASON_UNSPECIFIED = 0;
  //  {
  //    "code": 400,
  //    "reason": "X_X_X",
  //    "message": "..."
  //  }
  X_X_X = 1 [(errors.code) = 400];
}
`
