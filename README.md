Im here to write about my algorithm.
I'll give this passage to LLM's to continue developing my project.

Umbra is a ultra-secure messenger, made for the Privacy.
Literally everything is enciphered here. no one can read anything. everything is E2E encrypted.
So we can secure your privacy becuase "YOU MAKE YOUR PRIVACY"

You have your keys, you manage them and every transaction is decenterialized. 

We have "Server" just for a clean / secure connection to the Database. everything happenning is happenning in your device.
your device manages transactions, your device signs actions, your device decrypts and reads group chats, etc.

As I accidentally noticed just before, every single chat in our messenger is a "group chat".
You want direct messages? no problem, you can simply make a new group with only 2 members.

Let's talk about scenario. For example, let Alice, Bob, Carol and David be our users.
Every single user has a key pair. we call this key "usr_pub_key" (32 bytes) and "usr_priv_key" (32 bytes).
The Algorithm behind this key pair is "Curve25519"
Every action and transaction that will be done by this special key.
This key is actually the soul of the user. user is authorized if and only if he knows the "usr_priv_key".

"usr_priv_key" is stored in the server, but not as plaintext. it is stored ciphered with user's password.
We never store user's password. password is only used for generate the key of decrypting the ciphered private key.
We will call this ciphered private key:
	"iv_usr_priv_key" (12 bytes)
+	"enc_usr_priv_key" (32 bytes)
+	"tag_usr_priv_key" (16 bytes)
= 	packed together in this order: "ciphered_usr_priv_key" (60 bytes)
Note that every encryption algorithm used in Umbra is AES-256-GCM (excluding one exception when we have to use RSA-2048 encryption)

This is the scenario: Bob want's to login to system.
1. He generates a Key pair for his Session. The Algorithm behind this key is "RSA-2048".
We had to use RSA because we actually need a 1-side encryption for session starting.
Let's call this key pair: "sess_pub_key" (256 bytes) and "sess_priv_key" (1190 bytes)
2. Bob sends his "username" and "base64(sess_pub_key)" to the server
3. Server looks for Bob's username in the database. If it finds that, fetches the "ciphered_usr_priv_key".
4. Server encrypts "ciphered_usr_priv_key" again with Bob's session key ("sess_pub_key", RSA-2048)
And calls that "sess_ciphered_usr_priv_key"
5. Server Generates a random 16-byte string which is called "sess_token" and saves that for this specific session.
This is additional and final guarantee for user's session safety.
Server also encrypts this token to send that to the client with the same "sess_pub_key" which results "sess_sess_token".
As Server saved the "sess_token" to the database, it has it's uuid which should be called "sess_id"
6. Server responses the user with "base64(sess_ciphered_usr_priv_key)" and "base64(sess_sess_token)" and "base64(sess_id)"
7. Bob receives "sess_ciphered_usr_priv_key" and "sess_sess_token" and "sess_id"
8. Bob decrypts "sess_ciphered_usr_priv_key" with his "sess_priv_key" (RSA-2048) to find "ciphered_usr_priv_key"
And decrypts "sess_sess_token" with his "sess_priv_key" (RSA-2048) to find "sess_token"
9. Bob unpacks "ciphered_usr_priv_key" into "iv_usr_priv_key", "enc_usr_priv_key" and "tag_usr_priv_key"
10. Bob generates "key_usr_priv_key" which is derived from user's password
The Algorithm behind this derivation is Argon2ID (which needs an aditional 16-bytes random salt to be safe: "usr_psswd_salt")
I decided to use MemoryCost: 2 ** 19, TimeCost: 6 and Parallelism: 3 for this package that seems strong enough for this specific usage.
11. Now Bob can decrypt "enc_usr_priv_key" using this "key_usr_priv_key" to find his own "usr_priv_key"
12. Bob must store following properties in the Cookies:
	a. "usr_priv_key"
	b. "sess_token"
	c. "sess_id"
13. Whenever Bob will send any request to the server, he must hash the whole payload with key: "sess_token".
The Hash Algorithm behind this signing is HMAC-SHA-256 where its secret is the "sess_token".
Any request must hold "sess_id" as the main authorization key for the Server.
14. Whenever server reciecves any request from the client (including GET and POST methods), it must load the "sess_token" from the database based on received "sess_id" and validate user's payload with user's "sess_token".

Above scenario was only about Authorization process. I didn't still talked about message transfer protocols in Umbra.
So Let's start.
There are numerous Group chats available in the Umbra.
Any user might be member of a group chat, or invited to that at a moment.
So basically, as the soul of a group, any group chat has a secret key pair which is called "grp_pub_key" and "grp_priv_key"
The Algorithm behind this key pair is "Curve25519".
Obviously, "grp_pub_key" is very public and uses to generate the invition link, but "grp_priv_key" is the main difference between the users who are members of a group and other users.
Members of a group, all know this "grp_priv_key" and use that to communicate each other.
Group's key pair is not static and might change at some specific transactions, mostly when a new user joins the group or a user leaves it.
We call the process of key change "key_rotation".
This is necessary because for example when a member leaves the group, he must not access to new messages in the group. so messages must be encrypted with new keys.
As I noticed just before, all messages in a group are fully encrypted.
The encryption key for Group Messages are derived from "grp_pub_key" and "usr_priv_key" of the sender.
The Algorithm behind this is ECDH (Curve25519) but as the result of ECDH might not have enough entropy to be AES-256-GCM key, we will hash it.
The Hash Algorithm behind this derivation is Argon2ID (which needs an aditional 16-byte random salt to be safe: "msg_ecdh_salt")
I decided to use MemoryCost: 2 ** 16, TimeCost: 3 and Parallelism: 2 for this package that seems enough (we have many messages to be decrypted).
So Finally we have "msg_key" = Argon2ID(ECDH(grp_pub_key, usr_priv_key), msg_ecdh_salt)
Note that we don't have "ID" for messages. instead we use "UUIDv4" for each message that results a random 16-byte string which is so far the best option to be "msg_ecdh_salt". So we call messages' uuid's: "msg_ecdh_salt". They are literally the same.

Just like above, let's review another scenario: Alice want's to send a new message to a group where its memebers are herself, Bob and David.
1. Alice writes her message. message is utf-8 and it will turn into Buffer, and pad-end with 0 bytes until reach a power of 2. this is called "msg".
2. Alice knows her "usr_priv_key" and "grp_pub_key". So she use ECDH to make the "raw_shared_key" (32 bytes)
3. Alice generates a random uuid using UUIDv4 algo and uses that as "msg_ecdh_salt" in Argon2ID algo to generate "msg_key" (32 bytes)
4. She uses AES-256-GCM algorithm to encrypt her "msg" which results 3 binaries: "enc_msg" ("msg_len" := same length as "msg"), "iv_msg" (12 bytes) and "tag_msg" (16 bytes).
5. Alice concats "iv_msg" + "enc_msg" + "tag_msg" to a new {28 + "msg_len"}-byte blob which is called "ciphered_msg".
6. Alice uses her "usr_priv_key" to sign concat of "msg" and current unix seconds timestamp which is called "msg_time". result of signing is called "msg_signature".
7. Alice packs "ciphered_msg", "msg_ecdh_salt", "msg_time" and "msg_signature" together which is now called "msg_pack".
8. Alice signs the "msg_pack" with her "sess_token" in HMAC-SHA-256 algorithm to find "payload_signature".
9. She sends "base64(msg_pack)" and "base64(paylaod_signature)" and her "sess_id" to Server.
10. Server loads "msg_pack" and validates that with Alice's "sess_token" as I talked about it before.
11. As Server ensures everything is correct about Alice's request, it adds the payload to "validation_queue".
This Queue basically exists just to let other users validate sent messages rather than Server. this results more decenterialization of the Umbra.
12. As another user (for example Bob) load the service and checks the group for any new message, he automatically notices Alice's "msg_pack" which is waiting to be validated. so Bob will automatically decrypt the "ciphered_msg" into "msg" and validate that and "msg_time" using "msg_signature" and Alice's "usr_pub_key" and if verification done successfully, Bob will add a signature under Alice's "msg_pack" in "validation_queue".
Decryption process is explained at parts 14-16.
13. Everybody can visit pending requests in "validation_queue" and validate them. as at least 25% of members of the group (not considering Alice herself) validate her message, "validation_tag" sets to "1" and as at least 50% of members validate her message, "validation_tag" sets to "2".
The "validation_tag" is used to know how many people are validated a request. default is "0", "1" is half of the way and "2" means fully validated.
A Member (for example David) will see Alice's message if and only if at least one of those happened:
	a. David already validated her message
	b. her message reached tag "2"
If David already validated Alice's message but her message didn't reach tag "2", David can follow and know what level of "validation_tag" her message reached.
If anyone alerts that a "validation" failed, he/she will alert the Server and Server will recheck the "validation". If he/she was right about the alert, Server will set "validation_tag" to "-1" and call "Alice" to alert her: her message corrupted and now she has to decide what will she do.
14. When a member (for example Bob) wants to decrypt a message, he needs to reassemble "msg_key".
Obviously Bob starts by finding Alice's "usr_pub_key" and combine that with "grp_priv_key" using ECDH to find "raw_shared_key".
15. Bob will find "msg_key" using Argon2ID with salt = "msg_ecdh_salt".
16. Bob now unpacks "ciphered_msg" into "iv_msg", "enc_msg" and "tag_msg" and use them to decrypt "enc_msg" using AES-256-GCM into original "msg".
Bob now can validate "msg" and "msg_time" which is explained at parts 12-13.

Everything seems very secure so far but this is still not enough.
What if Server hacked and some unsafe data changed unexpectedly?
This is actually not good for decenterialization of the Umbra. right?
So, I decided to add a blockchain custom mechanism that logs every valid transaction across the server, which is used to ensure current data is all healthy and valid.
How it works? let me explain. Every single action on the Server (including sending messages, validating messages, alerting a message as an invalid message, changing messages' validity (="validation_tag"), joining new members, leaving a member, group's "key_rotation", or even group identical changes such as changing its name, etc) is a "transaction".
"transaction"s are stored in the blockchain chained together using their hashes.
The Hash Algorithm behind this chaining is SHA-256.
A Block holds related transactions (if any). which means for example if a user joins to the group, at the same moment we need a new group key.
So obviously joining transaction and group's "key_rotation" transaction are in the same block.
Just like that, when a member alerts the Server a message is invalid (which is a "transaction"), Server must check his/her alert at the same moment (which might cause a "validation_tag" change, which is another "transaction"). those "transaction"s are also will be in the same block.
Blocks use Merkle trees for their hashs, so they're very fast and reliable.
Blocks should be signed by members. Unlike the "validation_queue", if only 1 member signs the block, block is ready to be hashed and chained to the blockchain. We don't need multiple signs because blockchain is syncronous and should not stop the group until 25% or 50% of members accept (a) little transaction(s).
Obviously the signature of signer of the block should also considered when hashing ready block.
Every group has its own blockchain. so starting block of a group is its initialization and should be signed by its creator.
As you see, I used PoA (Proof of Authorization) for my blockchain, not PoW (Proof of Work) which needs difficult block minings and miners.
As a raw block inserted into the blockchain while it is waiting for signing, Server is looking for 1 member to ask him/her to sign the block. As any member checks the group, he/she will automatically ask Server if any pending block is there. If any, he/she will automatically sign that and send the signature to the Server. Then Server will replace the signature, compute hash of block and make it prooved in the blockchain.

Ok I know it's a little confusing now. So let me write you the third scenario.
Now, We have a complex scenario:
	a. Alice creates the group chat
	b. Bob joins the chat
	c. Alice says Hi to Bob
	d. Bob says Hi too
	e. Carol joins the chat too
	f. Bob says I have to leave, I'm sorry
	g. Bob leaves the chat
	h. Carol says Hi
	i. Mallory as the malicious character, finds the access to Carol's message while transfering to the server
	j. Mallory wants to change Carol's message and Server has to reject it
Just before starting, Note that Alice, Bob and Carol are authenticated users. So they have their "usr_priv_key".
Now, Let's start.
a:
	1. Alice wants to create a new group chat. Which means she is responsible with group's key: "grp_priv_key". So she simply generate a new keypair for the group.
	2. Alice chooses a "grp_name" and a small profile picure (max 512x512) for the group which is called "grp_prof_pic".
	If alice doesn't choose a profile picture, Umbra-Client is responsible to generate a totally random 512x512 colorful picture that needs to be accepted by Alice.
	3. Alice decides if the profile picture is public or it should be visible for members only. We always encrypt the profile picture however.
	4. The Decision of Alice, decides what would be the encryption password: based on "grp_pub_key" or "grp_priv_key".
	5. Whatever Alice decides, we will generate the "grp_prof_key" with Argon2ID (which needs an aditional 16-bytes random salt to be safe: "grp_prof_salt")
	I decided to use MemoryCost: 2 ** 16, TimeCost: 3 and Parallelism: 2 for this package that seems simple enough for this specific usage.
	So if Alice decides the profile picture should be private, then we will calculate "grp_prof_key" = Argon2ID(grp_priv_key, grp_prof_salt).
	Otherwise, if Alice decides the profile picture should be public, then "grp_prof_key" = Argon2ID(grp_priv_key, grp_prof_salt).
	As we need a uuid for the group, the best choice is let that be the same as "grp_prof_salt" which is generated using UUIDv4.
	6. As Alice found the encryption key: "grp_prof_key", now she is ready to encrypt the whole picture.
	The Algorithm behind this encryption is AES-256-GCM.