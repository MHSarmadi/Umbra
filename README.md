# Welcome to Umbra

Umbra is made to show you the **TRUE DIGITAL PRIVACY**.

This README is written about the architecture of **Umbra**. While you are reading it, you will figure out how does any part of Umbra work to ensure your privacy while messaging to anyone will never be in any danger.

> Now, Umbra is under developing. As soon as it released, you will be able to download Umbra and give it a test!

So. Let's start.

## Authorization and User Management

As you know, authorization of users is one of the most important parts of a safe service. Authenticated users are able to do anything simply, but others cannot do that theorically. How can we ensure this happens in practice?

This is where one of the Greatest parts of Umbra comes in: Umbra's Authorization.

Let's start by our famous digital characters: **Alice**, **Bob**, **Carol** and **David**.

Imagine Bob wants to keep his privacy safe. Obviously he has to `Encrypt` anything related to him. This is where we use our first algorithm: `AES-256-GCM` (Stands for `Advanced Encryption Standard, Key length: 256 bits = 32 bytes, Galois/Counter Mode`)

This algorithm is basically made to ask you a key (which is a 32-byte phrase) and `Encrypt`/`Decrypt` any digital content you want just by that special key. while you have it, you can either `Encrypt` or `Decrypt` anything. otherwise, you should not be able to know anything about a `ciphertext` which is already encrypted by the algorithm.

But... Bob is not going to write 32 `random` numbers (between 0 and 255) on a paper and enter it to the system as its password every single time he wants to log in the system, right? Alright, This is where our second algorithm comes in.

`Hash` algorithms are used to turn a text or any digital content into a fixed numbers of characters where they are not reversible! It means as you calculate the `hash` of a content, no one will be able to find your original content based on the `hash`. But everytime you calculate the `hash` of the same content, you will get the same result. This allows us to turn our `Human-Readable` passwords into an appropriate key with fixed numbers of characters (in our case 32 characters) to be used as `Encryption-Key`. We could simply use well-known `hash` functions such as `SHA-256` while they are already safe. But simple `hash` functions has some Cons so we will not use them for this case. One of the worst Cons is that such algorithms are not safe against `Brute-Force` attacks (where attacker guesses numerous `Human-Readable` passwords and `hash` them to check if it is the correct password). So new `hash` functions has been made to manually make the `hashing` process much much harder. In this way, guessing numerous passwords for attackers gets much much harder and almost impossible.

One of the best algorithms in this list is `Argon2` which has 3 modes for enhance its safety: `i`, `d` and `id` which means using both `i` and `d` together. I'll explain what do `i` and `d` mean later, For know however, let's use `Argon2id` which is the safest `hash` function for our purpose.

`Argon2` algorithm needs some options to work perfectly. They are: `TimeCost`, `MemoryCost`, `Parallelism`, `Version` and finnaly, `HashLength`. Let's break them all down:


|   Option   | What it does                                                                                                       | What we will use                |
| :-----------: | -------------------------------------------------------------------------------------------------------------------- | --------------------------------- |
|  TimeCost  | How many time the algorithm should run across the data? (More = Slower = Safer agaist`Brute-Force` attacks)        | Now: 16                         |
| MemoryCost | How much memory should algorithm occupy to do its calculations? (More = Slower = Safer against`GPU-Based` attacks) | Now: 2 ** 17 = 128 MiB          |
| Parallelism | How much thread should algorithm use                                                                               | 2                               |
|   Version   | Version of algorithm                                                                                               | 19                              |
| Hash Length | Length of the result                                                                                               | 32 (Key length of`AES-256-GCM`) |

There are also other features which are really common and helpful to enhance safty of `hash` functions which are adding `salt` and `pepper` to the main content. `salt` is a very `random` combination of characters (at least 16 bytes). Without this feature, if an attackers use a `Rainbow-Table` which is a table containing numerous usual passwords and their `hash`, he can find the password much faster. By adding (and obviously, saving) 16 `random` bytes called `salt`, we can make our `hash` result uniuqe and safe against `Rainbow-Tables`. `pepper` is a secret string which another safty layer and used when attacker has found user's password and the `salt` (from database) and wants to start `Decrypting` but he does not know the secret: `pepper`. `pepper` needs to be kept away from atackers. There are ways to keep it safe, but I decided to dynamically generate it in the client based on user's username and password. This way, user's username is also important in login process. This derivation uses `SHA-128` `hash` function which generates 16 `random` characters based on its input. The same `salt` which is used in previous `hasing` process is also used in this `hash` function. So let's give this architecture a summerize in codes:

```typescript
async function calculateUsersPrivateKeyDecryptionKey(username: string, password: string, salt: Buffer): Promise<Buffer> {
	const pepper = createHash("sha128").update(username).update(password).update(salt).digest()
	const timeCost = 2 ** 4, memoryCost = 2 ** 17, parallelism = 2 ** 1, version = 0x13, hashLength = 2 ** 5
	return await argon2.hash(password, {
		type: argon2.argon2id,
		salt,
		secret: pepper,
		timeCost,
		memoryCost,
		paralellism,
		version,
		hashLength,
		raw: true
	})
}
```

Note that the point of using `Argon2` algorithm is this algorithm is slow. So we have to use it asyncronous.

Also, `SHA-128` stands for `Secure Hash Algorithm, Output length: 128 bits = 16 bytes`

So, as Bob successfuly found his `Decryption-Key`, has to `Decrypt` his own `Private-Key` which is his soul across Umbra. Let me explain.

Some algorithms are made to generate `Key-Pair`s which means that you generate yourself 2 keys: `Public-Key` and `Private-Key`. The `Public-Key` is a **public** combination of bytes which is used by other users to interact with you securely, While you only one who knows your `Private-Key`. Imagine you want to send everyone a message and you want to digitally `Sign` it, so you can proof anyone that you yourself sent this message not a malicious user like Mallory! So you can somehow `Sign` your message with your own `Private-Key`, and anyone can `Validate` it using your `Public-Key`! This way, no one but you could generate this `Signature`, however anyone can proof that you have sent the message yourself using your `Public-Key`. This seems great! Isn't it?

So, there are basically two famous algorithms that generate `Key-Pair`s in front of us: `RSA` and `Elliptic Curves (EC)`. `RSA` is the older algorithm and uses significantly larger keys (so its slower) and funny enough it does not even offer more safety, so we are not going to use that now (Somewhere else we have to use it). So let's use `Elliptic Curve Cryptography (ECC)` as our main algorithm. But there are different `Curves` to be used for calculations, So we have to choose one of them. Based on my researches, one of the best and the fastest and the safest `Curves` to be used is `Curve25519` which is well-known across cryptography world.

So, Bob and anybody else already generated a `Key-Pair` based on `Curve25519` for themselves when they wanted to sign up the Umbra. They already `Encrypted` their `Private-Key` using `AES-256-GCM` algorithm with key = the `Decryption-Key` we just calculated and saved it in the Database of Umbra. Note that `AES-256-GCM` `Encryption` algorithm also needs two important values: `iv` (12 bytes) which is generated by Bob himself and `tag` (16 bytes) which is generated by the algorithm.

As the `Public-Key` is obviously public, they saved their `Public-Keys` purely in the database. Just to ensure the `Public-Key` was not altered by any attacker, they already `Signed` their `Public-Key` with their `Private-Key` and saved the result in the database. So if anyone hurts the `Public-Key` or its `Safety` signature, anyone will be noticed and user himself tries to cure that next time he logs in the Umbra.

This is a summerize of this `Decryption` process:

```typescript
function decryptPrivateKey(decryptionKey: Buffer, iv: Buffer, encryptedPrivateKey: Buffer, tag: Buffer, currentPublicKey: Buffer, currentPublicKeySignature: Buffer): Buffer {
	const decipher = createDecipheriv("aes-256-gcm", decryptionKey, iv).setAuthTag(tag)
	const privateKey = Buffer.concat([
		decipher.update(encryptedPrivateKey),
		decipher.final()
	])
	// Assert: The Result Of Public Key Derivation of The Decrypted Private Key Must Be Exactly The Same As currentPublicKey
	// Assert: The currentPublicKeySignature Must Be Valid Against currentPublicKey
	return privateKey
}
```

> There might be a B-Plan for users who lost their passwords or their `Encrypted` `Private-Key` on the Database is corrupted. I don't plan to develop in production this feature in 1.0.0 release, I have the ideas however. I will explain it at the end of this README.

While now, Bob saved his `Encrypted` `Private-Key` and its `iv` + `tag`, his raw `Public-Key` and its `Signature` and the `salt` for generating `Encryption-Key`. Now, he wants to log in. This means now he wants to `Decrypt` his own `Private-Key` from the server. He requests the server to find those saved values based on his `username`. Server responses him his `Encrypted` `Private-Key` and its `iv` + `tag`, his raw `Public-Key` and its `Signature`, the `salt` and finally a new `random` value called `Session-Token` and its `Expiry-Time`. Then, Bob starts by calculating the `Decryption-Key`using his`username`, `password` and the`salt`just as we talked before. Then, he tries to `Decrypt`the`Encrypted` `Private-Key`with its`iv`, its `tag`and the calculated `Decryption-Key`. As we are using `GCM` mode of`AES-256`rather than other modes and saving`tag` of `Encryptions`, we sure that if we decrypted the `Private-Key` successfully, its healthy and not corropted. just to make sure`Public-Key`and its`Signature`are healthy too, Bob once derive his`Public-Key`from his `Decrypted` `Private-Key`and checks if it is the same as loaded `Public-Key`. Finally he `validates`the`Signature`using his`Public-Key` to make sure everything is healthy.

As I noticed once in the previous paragraph, Server responses not only saved `values` related to Bob's `Private-Key` and its `Encryption`, it responses him a new `random` 16-byte string called `Session-Token`. Once the server receives the Login request, it generates a `random` string and saves that as `Session-Token` of the user in the Database. User (Bob) receives that and keep that safe. But what is it used for? I once said the personal `Private-Key` of a user is his/her soul in Umbra, which means whoever has it, we trust him/her. But an attacker might find it however (with a very very low probablity). So how can we make sure a user is himself/herself not an attacker? a simple way is to `Sign` any request before sending it using the `Session-Token`. This way, only users who asked server a `Session-Token` are allowed to send Server new Authorized requests. Also, an attacker might try the `Replay-Attack` (Which means to record a valid request across the network and send the same later). So user must always generate another 32-byte `random` string which is called `nonce` and joins that to his request and `Sign` that using `Session-Token` too. This way, server once receives a request with a specific `nonce`, adds the `nonce` to the `Black-List` of `nonces` so no attacker can use the same `nonce` again.

Obviously, Bob should save his `Private-Key` and the `Session-Token` after a successful login. But how many times can Bob use the same `Session-Token`? Well I said before that the Server gives Bob a `Session-Expiry-Time` (Based on successful logins of Bob, more = longer) and Bob has to plan for update his `Session-Token` regularly. But how to make sure Bob himself is asking server a new `Session-Token`, not an attacker? a simply way is to `Sign` this request twice: once with the expired `Session-Token`, once with his `Private-Key`. But it seems just a little weak. How to reinforce this process?

Bob had to generate another `Key-Pair` before sending the Login request, this time with `RSA-2048` algorithm! Why? Let me explain. Besides of all Pros of `ECC`, `RSA` offers another feature that `ECC` can't which is `1-side Encryption`. This means while anyone has the `RSA-Public-Key` of someone, they can `Encrypt` a content for him/her. But no one can `Decrypt` and understand the content but himself/herself who has the `RSA-Private-Key`. We're going to use it as well. Bob generates a `RSA-Key-Pair` called `Session-Key-Pair` and sends the server besides of his username. Server will save his `Session-Public-Key` and use that for every single response. In fact, ALL of Authorized responses of the server must be fully `Encrypted` with this `Session-Public-Key` including the new `Session-Token` while updating it. This way, Bob MUST once give the server a `Session-Public-Key` to be able to ask it a new `Session-Token` after its `Expiry-Time`. Combination of including old `Session-Token` and Bob's main `Private-Key` (using `Signing`) in the request and receiving new `Session-Token` `Encrypted` using the `Session-Key-Pair` makes us strongly sure no attacker can fake his identity using `Session-Tokens`. By the way, a `Session-Key-Pair` has no `Expiry-Time` so once the user losses the `Session-Private-Key`, its `Expiry` has came. Obviously Bob had to save his `Session-Private-Key` as well as his main `Private-Key` and the `Session-Token`.

By the way, Bob has to tell the server which session he has. We're simply using cookies to store a `Session-UUID` which tells the server which `Session-Token` and `Session-Public-Key` is related to user. But don't worry. Guessing random 32-byte `UUIDs` is not going to help attackers with `Session-Attack`, because simply they don't have accurate `Session-Private-Key`. So they will never understand the responses. Never forget an attacker can never `Sign` his requests with the `Session-Token` within its `Expiry-Time` too. Not going to lie, its extermly safe now, isn't it?
