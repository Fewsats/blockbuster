# Blockbuster: Seamless Video Monetization with Lightning Network

[Blockbuster](https://blockbuster.fewsats.com) is a video sharing platform that allows content creators to monetize their videos instantly using the L402 protocol & the Lightning Network. 

With just an email sign-up, creators can upload and start earning, while viewers can purchase access using their public keys without creating accounts.

You can use blockbuster at https://blockbuster.fewsats.com

## How It Works

![Blockbuster Overview](https://example.com/blockbuster_overview.png)

1. **Content Creators** upload videos and get L402 URIs
2. **Viewers** use public keys (e.g., Nostr ID) to purchase access
3. **L402 & Lightning Network** facilitates instant, low-cost payments


### Video Upload and L402 URI Generation

When a content creator uploads a video, the system:
1. Uploads the video to storage
2. Generates a unique L402 URI for the video
3. Stores video metadata and L402 URI in the database

#### L402 URI Scheme

An L402 URI is generated for each uploaded video. It follows this scheme:

```
l402://blockbuster.fewsats.com/video/info/79c816f77fdc4e66b8cd18ad67537936
```

and contains the following information in JSON format.

```json
{
  "version": "1.0",
  "name": "How to focus and think deeply | Andrew Huberman and Lex Fridman",
  "description": "Lex Fridman Podcast full episode:  ...",
  "cover_url": "https://pub-xxxerea.r2.dev/cover-images/79c816f77fdc4e66b8cd18ad67sfas537936",
  "content_type": "video",
  "pricing": [
    {
      "amount": 1,
      "currency": "USD"
    }
  ],
  "access": {
   "endpoint": "https://blockbuster.fewsats.com/video/stream/79c816f77fdc4e66b8cd18ad67537836",
    "method": "POST",
    "authentication": {
      "protocol": "L402",
      "header": "Authorization",
      "format": "L402 {credentials}:{proof_of_payment}"
    }
  }
}
```

Creators can share this URI, which provides information about the content and the endpoint where you can pay for it.

### Buy video with L402 URI

The process of purchasing access to a video using an L402 URI involves several steps:

1. **Fetch Video Info**: 
   The viewer's client fetches video metadata through the L402 URI.

   ```
   GET https://blockbuster.fewsats.com/video/info/79c816f77fdc4e66b8cd18ad67537936
   ```

   Server responds with JSON:

   ```json
   {
     "version": "1.0",
     "name": "How to focus and think deeply | Andrew Huberman and Lex Fridman",
     "description": "Lex Fridman Podcast full episode: ...",
     "cover_url": "https://pub-xxxerea.r2.dev/cover-images/79c816f77fdc4e66b8cd18ad67sfas537936",
     "content_type": "video",
     "pricing": [
       {
         "amount": 100,
         "currency": "USD"
       }
     ],
     "access": {
       "endpoint": "https://blockbuster.fewsats.com/video/stream/79c816f77fdc4e66b8cd18ad67537836",
       "method": "POST",
       "authentication": {
         "protocol": "L402",
         "header": "Authorization",
         "format": "L402 {credentials}:{proof_of_payment}"
       }
     }
   }
   ```

2. **Request Payment Challenge**:
   
The user sends request payment to the L402 `endpoint`, with their `public key`, `domain` (ie. `blockbuster.fewsats.com`), current `timestamp` and a signed `domain:timestamp` message . 

```
POST https://blockbuster.fewsats.com/video/stream/79c816f77fdc4e66b8cd18ad67537836
{
    "pub_key": "03...public_key...",
    "domain": "blockbuster.fewsats.com",
    "timestamp": 1686123456,
    "signature": "304...signature..."
}
```

The signature and public key serve multiple purposes:

1. Proves that the request comes from private key owner.
2. Prevents unauthorized access and tampering with the request.
3. Protects against replay attacks by including a timestamp.

This approach allows for authentication without requiring traditional user accounts and previous sign up.



3. **Server Processing**:
   The server handles the request as follows:

   ```go
   func StreamVideo(request) {
     // Validate the signature
     if !ValidateSignature(request.PubKey, request.Signature, request.Domain, request.Timestamp) {
       return ErrorResponse("Invalid signature")
     }

     // Create L402 challenge
     challenge = CreateL402Challenge(request.PubKey, videoExternalID)

     // Respond with 402 Payment Required status and challenge
     return PaymentRequiredResponse(challenge)
   }
   ```

    The `402 Payment Required` challenge is included in the `Www-Authenticate` header and contains
    the Lightning Invoide and a unique macaroon tied to the user's public key:

   ```
   HTTP/1.1 402 Payment Required
   Www-Authenticate: L402 macaroon="AGIAJEemVQUTEyNCR0exk7ek90Cg==", invoice="lnbc1500n1..."
   ```

4. **Paying the Invoice**:
   The viewer pays the Lightning invoice using their preferred wallet. 
   
   After payment, the viewer sends another request including the macaroon and preimage:

   ```
   POST https://blockbuster.fewsats.com/video/stream/79c816f77fdc4e66b8cd18ad67537836
   Authorization: L402 macaroon:preimage
   ```

5. **Server Verification and Response**:
   The server verifies the credentials and responds with the video URL:

   ```go
   func StreamVideo(request) {
     paymentHash = ValidateL402Credentials(request.AuthorizationHeader)
     if paymentHash == "" {
       return ErrorResponse("Invalid credentials")
     }

     RecordPurchaseAndView(videoExternalID, paymentHash)
     streamURLs = GenerateStreamURL(videoExternalID)
     return SuccessResponse(streamURLs)
   }
   ```

    The response include 2 URLs to stream the video either using HLS or Dash formats. URLs can be used either in browser-based or desktop video players.
    
   ```json
   {
     "hls_url": "https://videodelivery.net/9876543210abcdef/manifest/video.m3u8",
     "dash_url": "https://videodelivery.net/9876543210abcdef/manifest/video.mpd"
   }
   ```


This flow ensures secure, account-less authentication and payment for video access using the L402 protocol and Lightning Network.

## System Architecture

Blockbuster is written in golang and built with a modular architecture. The main three modules are `video`, `auth` and `l402`.

### Video

Video module handles:

* upload video proccess `/video/upload`
* L402-protected stream video `/video/stream/:id`
* L402 URI video info `/video/info/:id`
* list user videos `/user/videos`

### Auth

Auth module is responsible of email authentication system for content creators.
    * Email login link `/auth/login`, `/auth/logout`
    * User info `/me`
* Viewers using a `public key` using auth middleware [l402/authenticator.go](l402/authenticator.go)


### L402

The L402 module defines an authenticator [l402/authenticator.go](l402/authenticator.go) that is responsible for:
* Protecting videos behind a L402 paywall
* Minting macaroons tied to the viewer `public key`
* Verifying the signature sent on the payment requests


## Configuration and Deployment

The system is highly configurable through the [sample-config.conf](sample-config.conf) file. Key settings include:

- Database connection
- Storage service
- Lightning Network provider
- L402 protocol settings

## Run the server

1. Clone the repository
2. Copy `sample-config.conf` into `blockbuster.conf` and configure:
    - Database connection (defaults to SQLite)
    - Storage service
    - Lightning Network provider
    - L402 protocol settings
3. Run the server:
   ```
   go run cmd/server/main.go
   ```


## Contributing

We welcome contributions!

## License

Blockbuster is released under the [MIT License](LICENSE).