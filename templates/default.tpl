proxy_http_version 1.1;
proxy_buffering off;
proxy_set_header Host $http_host;
proxy_set_header Upgrade $http_upgrade;
proxy_set_header X-Real-IP $remote_addr;
proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
# Mitigate httpoxy attack
proxy_set_header Proxy "";

client_max_body_size        10m;
client_body_buffer_size     128k;
client_body_temp_path       /var/tmp/;

proxy_connect_timeout       90;
proxy_send_timeout          90;
proxy_read_timeout          90;


server {
    listen      80;
    server_name "";
    return      444;
}

server {
    listen      443;
    server_name "";
    return      444;
}


{{$p := .Certbot}}
{{ range .Domains}}
    {{ if gt (len $p) 0 }}
        server{
            server_name {{.Domain}};
            listen 80;
            {{ if .Encryption }}
                location /.well-known/acme-challenge/ {
                    auth_basic off;
                    allow all;
                    root /usr/share/nginx/html;
                    try_files $uri =404;
                    break;
                }
                location / {
                    return 301 https://$host$request_uri;
                }
            {{ else }}
                location / {
                    proxy_pass http://{{.Proxy}};
                }
            {{ end }}

            access_log /var/log/nginx/access.log main;
        }
        {{ if .SslReady }}
            server {
                server_name {{.Domain}};
                listen 443 ssl http2;

                ssl_certificate /etc/letsencrypt/live/{{.Domain}}/fullchain.pem;
                ssl_certificate_key /etc/letsencrypt/live/{{.Domain}}/privkey.pem;
                include /etc/letsencrypt/options-ssl-nginx.conf;
                ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;

                include /etc/nginx/vhost.d/{{.Domain}}*.conf;
                location / {
                        proxy_pass http://{{.Proxy}};
                }

                if ($scheme != "https") {
                        return 301 https://$host$request_uri;
                } # managed by Certbot

                access_log /var/log/nginx/access.log main;
            }
        {{ end }}

    {{ else }}
        server{
            server_name {{.Domain}};
            listen 80;

            location / {
                proxy_pass http://{{.Proxy}};
            }

            access_log /var/log/nginx/access.log main;
        }
    {{ end }}
{{end}}

