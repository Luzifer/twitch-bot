+++
title = "Preparations"
weight = 2
+++

{{< lead >}}
In order to communicate with Twitch you need to set up an application in the Twitch developers console.
{{< /lead >}}

{{< alert style="warning" >}}
In case you are setting up multiple Twitch-Bot instances, create **one application per instance**! Re-using an existing application even for a test-instance will lead to unexpected results like randomly breaking bot-authorizations!
{{< /alert >}}

Registering your application is a relatively straight-forward process:

- Go to https://dev.twitch.tv/console/apps/create
- Fill out the form you are presented with. You can choose any **Name** you want for your bot. I'd recommend using one you later will recognize your bot under. For the **OAuth Redirect URL** choose the URL you want to have the bot available under later. If you want to have it running locally you can choose `http://localhost:3000/` for this field.  
  ![](/screen-twitch-console-register-app.png)  
- After registering your application go into the application you've just created and click the **New Secret** button. Note down the **Client-Id** and **Client-Secret** in a safe place. You will need them in the [Configuration]({{< ref "configuration.md" >}}) step.
