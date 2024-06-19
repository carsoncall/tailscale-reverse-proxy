### Purpose

I am a huge fan of Tailscale. I find it to be one of the most exciting developments in networking, and it is the the backbone of my home lab. Being double NAT-ed, I had no other choice. I have an old machine that I use to host most of my services, such as Jellyfin, Immich, and HomeAssistant; however, Tailscale doesn't really support reverse proxies on Tailnets. Anyway, this project uses the fantastic tsnet library published by Tailscale to create virtual nodes, each in its own thread, to forward traffic to specific places on your network. You can use this to forward traffic to locations on your LAN or on your Tailnet.

### Getting Started
There are a couple of things you need to get started:
1. proxy.conf
   This is a very simple key-value pair file of what you want the proxies to do. The key is the name of the node (the "hostname") that will appear on your tailnet, and the value is the URL you want it to forward to. Check out the included proxy.conf, which is the configuration that I use personally.
2. .env file
  This is where you will keep your Tailscale secret key. It should be TS_AUTHKEY=tskey-auth-******. You can create one of these by logging in, going to Settings, and then Create New Key. There are a couple of different options for these; you probably want to make them Reusable and Ephemeral, so that you can create multiple nodes multiple times, and they disappear when the program isn't running. 
3. go run .
   (EDIT): there appears to be an issue with one of the dependencies of tsnet that causes Go to fail to resolve the dependencies. I'll update this once it's fixed and I have time to figure out what is going on. 
