# Bevelgacom WAP Portal

We are Bevelgacom a non-profit hobby run retro ISP focussing on keeping old devices and their technologies alive and useful in the modern age.
Our WAP website at `wap.bevelgacom.be` aims to revive the old mobile provider style WAP portals. Supplying a mix of out own services as linking in our portal to others. 

Our aim in this project is to provide useful services to WAP 1.0-2.0 phones! 

Our current services:

- News from [VRT NWS](https://vrt.be/nws) (Ducth)
- DB Navigator for most European Trains!
- Barcode generator

## Portal links

Do you run a WAP website? We need you! We are tryng to collect a portal of all WAP sites still available.

What is required to apply?

1. Your website exists. (Seems silly but we live in an age of genrative AI nonsense)
1. You support the Wireless Markup Language (Currently we do not accept cHTML, compact xHTML etc websites)
1. (OPTIONAL but will get higher ranking) Your website accomodates older WAP phones in size limits, WBMP images, ...

Your website applies? GREAT! Feel free to PR yourself in `static/portal-*.wml`. Thanks a lot!

## I want to run a WAP site? Where do I start

Hello friend! Welcome to a journey like none else! Here is a quick start guide to get you going!

A few lessons:
- Take inspiration from others, our repo is there for you to copy!
- Forget the internet, forget AI. They might mislead you. Trust old resources.
- Your only friend is [Archive.org](https://archive.org)

### Books

- [Getting started with WAP and WML](https://archive.org/details/gettingstartedwi0000evan/) (TIP: skip the part on Java, any modern server side language works)

### Hardware

Want to get to run it on real hardware? Here is a small guide what to look out for:

- Any modern HMD Global produced Nokia feature phone was WAP support. Exception are those based on KaiOS.
- Most phones (not the  Nokie1xxx budget series) between 2002 and 2007 have perfect WAP support, post iPhone age this became different.
- Avoid phones only supporting WAP over CSD (modem calls), nearly 99% of 2G networks no longer support these calls even to your own modem.
- Look for a phone with GPRS/EDGE support, many security critical infrastructure still uses GPRS your carrier will support it for years to come even longer than 3G.
- Ask your family and friends if they have old phones! Think Green!
- Look for popular phones back of their age like the Nokia 3510i (GRPS support + color) you will find one on every fleemarket
- Old is fun! If you have a WAP 1.x phone you need a proxy to convert WML to WML Binary encoded pages. [Bevelgacom](http://bevelgacom.be) hosts a public one based on [Kannel](http://kannel.org)

### Emulators

Emulators of WAP mostly died of old age, they all are a bit crappy but will work:

- [Nokia 3330 simulator by Nokia](https://archive.org/details/3330_Simulator)
- [wApua](https://fsinfo.noone.org/~abe/wApua/Download.html) NOTE: very limited but will render WBMP and basic text, buttons. No input fields sadly.

### Validators

WAP proxies are very strict on the XML being valid as they convert them into way smaller binary formats for transmission. Validate your output before trying the website out.

- [W3C](https://validator.w3.org/)

### Hosting

Any webhost will do! Just make sure that **HTTPS is OPTIONAL or DISABLED**. WAP will not support secure TLS standards.

Note you might have to add a few MIME rules

* `text/vnd.wap.wml` for WML
* `image/vnd.wap.wbmp` for WBMP

## Public GSM Provider capabilites

(PRs welcome from your own testing)

### Belgium

| Operator          | CSD  | WAP over SMS | GPRS | EDGE |
| :---------------- | :--: | :---------:  | :--: | :---:|
| Orange| X | OK* | OK | OK
| Proximus | X | OK** | issues on 5G towers | ~ ***
| BASE/Telenet | X | GSM bands not compatible with Nokia 7110 | OK  | OK
| Hey (Orange MVNO) | X | OK* | OK | OK
| Mobile Vikings (Proximus) | X | OK** | issues on 5G towers | ~***
| Digi ****| X | X | X | X
| Neibo (Orange MVNO)| X | OK* | OK | OK

`* The Mobistar network seems to need an explicit SMSC set, it will drop deliveries of the WAP browser if the "Server Number" is not set to "+32 495 002 530"`  
`** Needs 3rd party gateway such as Bevelgacom`  
`*** Proximus shows issues on 2G data, milage may vary`  
`**** Digi shows total failure to register on a 2G network, has issues with SMS centers being coded as Romania on SIM. And various other hasty deployment issues, would have loved to give them a better score but this is the sad truth, never use them for retro phones`

### The Netherlands

| Operator          | CSD  | WAP over SMS | GPRS | EDGE |
| :---------------- | :--: | :---------:  | :--: | :---:|
| Vodafone | OK* | Untested| OK | OK
| Odido | X | X | X | 2G shutdown


`*` CSD to our suprise works! Set to Analog and use a [Dial up](https://bevelgacom.be/products/dial-up/) number. This is tested 02/2025 as CSD has been phased out everywhere else we have no guarantees of this.


### Austria

| Operator          | CSD  | WAP over SMS | GPRS | EDGE |
| :---------------- | :--: | :---------:  | :--: | :---:|
| Yesss (A1) | X | ? | OK | ?


## How to acheive maximum compatibility, aka the Nokia 7110 standard

You want to get to the true 1999 standard? Perfect! The first major selling phone with WAP was the Nokia 7110. It did have a few quirks:

* Tables only got added in the last firmware
* Caching controls is broken, it will always cache think abour unique URLs if needed
* Long URLs are cut of at ~100 bytes in SMS mode, absolute max is 100 for home, 255 for others
* 1.3 kilobytes binary data per card deck maximum
* It has no back button, really! You need to program one in every card

```xml
<do type="prev" label="Back">
<prev/>
</do>
```