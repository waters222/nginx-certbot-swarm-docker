proxy_http_version 1.1;
proxy_buffering off;
proxy_set_header Host $http_host;
proxy_set_header Upgrade $http_upgrade;
proxy_set_header Connection $proxy_connection;
proxy_set_header X-Real-IP $remote_addr;
proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
proxy_set_header X-Forwarded-Proto $proxy_x_forwarded_proto;
proxy_set_header X-Forwarded-Ssl $proxy_x_forwarded_ssl;
proxy_set_header X-Forwarded-Port $proxy_x_forwarded_port;
# Mitigate httpoxy attack (see README for details)
proxy_set_header Proxy "";

client_max_body_size        10m;
client_body_buffer_size     128k;
client_body_temp_path       /var/tmp/;

proxy_connect_timeout       90;
proxy_send_timeout          90;
proxy_read_timeout          90;


{{range .Domains}}
    {{if len .Certbot | eq 0}}
        server{
            server_name {{.Domain}};
            listen 80;

            location / {
                proxy_pass {{.Proxy}};
            }

            access_log /var/log/nginx/access.log vhost;
        }
    {{else}}
        server{
            server_name {{.Domain}};
            listen 80;
            location /.well-known/acme-challenge/ {
                proxy_pass {{.Certbot}};
                break;
            }

            return 301 https://$host$request_uri;

            access_log /var/log/nginx/access.log vhost;
        }

        server {
            server_name {{.Domain}};
            listen 443 ssl http2;

            ssl_certificate /etc/letsencrypt/live/{{.Domain}}/fullchain.pem;
            ssl_certificate_key /etc/letsencrypt/live/{{.Domain}}/privkey.pem;
            include /etc/letsencrypt/options-ssl-nginx.conf;
            ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;

            include /etc/nginx/vhost.d/{{.Domain}}.conf;
            location / {
                    proxy_pass {{.Proxy}}
            }

            if ($scheme != "https") {
                    return 301 https://$host$request_uri;
            } # managed by Certbot

            access_log /var/log/nginx/access.log vhost;
        }
    {{end}}
{{end}}

