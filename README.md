Telegram TDLib Giveawayer
=========================

# Introduction
Make automatic, periodic giveaways via @MoneroTipBot on Telegram.

This is not a Telegram bot. It uses [tdlib](https://core.telegram.org/tdlib) and works under your username. Let it run on a VPS and make unattended giveaways.

# Get Your Telegram API Keys
Visit https://my.telegram.org/ and register this app for your Telegram account.

# Usage
**Note**: The first time you run this command it will prompt your for your phone number and password (if you have set a cloud password in Telegram). The phone number format must be like this: `+49171123456789` if the phone number was `0171123456789`.

Run the command without any arguments to get a list of all the groups you are a member of.

```sh
go run main.go
```

Now open a file and write down the group IDs you want to make a giveaway line by line and add the number of giveaways you want to make in each group, seperated by a colon, like this:

```sh
-123456789:1
```

It means, for group ID `-123456789` one giveaway shall be made.

Save the file and run the app with the appropriate arguments:


```sh
go run main.go -apiID 123456 -apiHASH MYLONGAPIHASH -filename mygroups.txt -giveaways 3 -every 12h
```

Note: the `-every` command has to be in this format: `h` (for hours), `d` (for days), `m` (for minutes), etc.


Happy Tipping!

# Contribution
* You can fork this, extend it and contribute back.
* You can contribute with pull requests.

# Donations
I love Monero (XMR) and building applications for and on top of Monero.

You can make me happy by donating Monero to the following address:

```
89woiq9b5byQ89SsUL4Bd66MNfReBrTwNEDk9GoacgESjfiGnLSZjTD5x7CcUZba4PBbE3gUJRQyLWD4Akz8554DR4Lcyoj
```

# LICENSE
MIT License