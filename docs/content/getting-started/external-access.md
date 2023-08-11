+++
title = "External Access"
weight = 4
+++

{{< lead >}}
In order to be able to access the bot through a web-browser and make configuration changes using the web-interface we need to securely expose the interface to the internet.
{{< /lead >}}

{{< alert style="info" >}}
In case you did the installation of the bot on your **local machine**, skip this part. You can access the web-interface at `http://localhost:3000/` after you've continued with the [Configuration]({{< ref "configuration.md" >}}).
{{< /alert >}}

In order not to make this a quite long and extensive tutorial we'll use two tutorials of DigitalOcean to aid us:

- [DigitalOcean: How To Install Nginx on Ubuntu 20.04](https://www.digitalocean.com/community/tutorials/how-to-install-nginx-on-ubuntu-20-04)
- [DigitalOcean: How To Secure Nginx with Let's Encrypt on Ubuntu 20.04](https://www.digitalocean.com/community/tutorials/how-to-secure-nginx-with-let-s-encrypt-on-ubuntu-20-04)

We will follow the first one up to step 4, and omit step 5 and 6. These are not required for us as they are configuring a locally hosted website which we don't want to do for the bot. Instead of the suggested `/etc/nginx/sites-available/your_domain` file in the first tutorial we will create a `/etc/nginx/sites-available/twitch-bot.conf` file with the following content:

```nginx
server {
  listen 80;
  listen [::]:80;

  server_name twitch-bot.mydomain.com;

  location / {
    add_header X-Robots-Tag noindex;

    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "Upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    proxy_pass http://127.0.0.1:3000/;
  }
}
```

Make sure to replace the `server_name` directive with the (sub-)domain you want the bot to be available at.

After creating the file we'll enable it similar to the tutorial:

```console
# Link the configuration to the enabled ones
$ sudo ln -s /etc/nginx/sites-available/twitch-bot.conf /etc/nginx/sites-enabled/

# Check whether there are any errors
$ sudo nginx -t

# If there were no errors, restart the nginx service
$ sudo systemctl restart nginx
```
