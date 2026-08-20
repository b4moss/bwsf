# Development and Design Philosophy

If you arrived here looking for the `bwsf` command, this content might not be necessary for you.

However, for those who aren't familiar, we'll document here what concept `bwsf` is based on and what motivated its development.

## What Are Environment Variables?

Before discussing `bwsf` and `.env` files, you need to understand what environment variables are.

Environment variables are **variables maintained by the operating system**.

You might be thinking, "What does that even mean?"

A **variable**, in computers and programming languages, is a **container for values**.

For example, in the PHP programming language:

```php
$value = "Hello world";
```

This means the value "value" (the $ is a marker for variables) contains the string "Hello world" (assignment).

Therefore, when you execute this with PHP's `echo` command:

```php
echo $value;
// -> Hello world
```

"Hello world" is output.

The above is just a PHP example, and the syntax differs between programming languages, but the concept of variables is universal in the computer world.

Environment variables are not specific to programming languages—they're handled by the OS itself. While there are various operating systems, on Linux and macOS:

```bash
export VALUE="Hello world"
echo $VALUE
# -> Hello world
```

This produces the output shown.

Environment variables have the following advantages and disadvantages:

- **Advantage**: You can conceal content using variable names
- **Disadvantage**: Since they're stored in memory, variable values are reset when you restart

These are the characteristics.

## The Twelve-Factor App

There's a concept called [The Twelve-Factor App](https://12factor.net/).

This is essentially a collection of best practices that have been shaped through years of web application development.

For example, Ruby on Rails broadly bases itself on this philosophy.

Ruby on Rails and other frameworks that inherit Rails' philosophy (like Laravel and Django) are also based on this thinking.

These so-called "Rails way" frameworks have been adopted by many products and operate stably.

In other words, The Twelve-Factor App is a very important way of thinking for developing robust web applications.

As the name suggests, it actually consists of **12 factors**.

The third factor is **[Config - Store config in the environment](https://12factor.net/config)**.

While I encourage you to read the link for details, simply put:

- Database connection URLs
- Usernames and passwords
- API keys, etc.

This information should be managed using environment variables.

For example, if you write this in a configuration file like `config.rb`, it will be recorded by **VCS (Version Control System, with Git being the de facto standard today)**, which is indispensable in current application development.

Even if you're not using a public repository like GitHub, there's a risk that configuration values, passwords, and other **sensitive information** will remain in the version control system, be shared with other developers, and leak from there.

Therefore, The Twelve-Factor App established the principle that **configuration values should be stored in environment variables**.

## The dotenv Approach

However, as mentioned earlier, environment variables have unwieldy aspects.

You can't see the contents without outputting with the `echo` command, and you can't easily tell what environment variable names (key names) exist in the terminal.

There is a `printenv` command that can list them all, but conversely, it outputs all environment variables across the entire terminal, which can be unwieldy.

Furthermore, since the information disappears when the terminal is restarted, developing with environment variables has various hassles.

This is where the `dotenv` approach was born.

You place a `.env` file at the root of your development directory.

Next, you install a package called `dotenv` for your development language. These have been ported to various languages and are available for PHP/Node.js/Python/Ruby/Go/Docker, etc.

The `dotenv` package reads the .env file at your project root and exports the information written there as environment variables.

A typical .env file looks like:

```text
DATABASE_URL=postgres://user:password@db.example.com:5432
DATABASE_USER=john
DATABASE_PASSWORD=secretpassword
```

This is similar to what you'd write when using `export` on the command line.

**A very important point** is that these **.env files should be excluded from version control systems**.

Because if you don't, configuration values will be version-controlled just like in the `config.rb` case.

In other words, it's important that .env files exist only on users' local machines.

With Git, you can exclude files from Git management by adding the following to your `.gitignore` file:

```text
.env
```

## Multiple Developers Are the Norm

Now, if you're developing alone, you can just save the .env file in your directory, but in modern software development, it's normal to work with multiple people.

This creates the need to share these .env files.

**But wait a moment.**

Some of you may have already had this experience.

For example, have you ever "**shared .env files via email/chat tools like Slack/file sharing tools like Dropbox or Google Drive**"?

This is a **very dangerous practice**.

From The Twelve-Factor App perspective, configuration values are information that **must be kept secret**.

However, with email they pile up in your inbox, and the same goes for Slack.

What about Dropbox or Google Drive? You might delete them later, but they remain as revisions.

Are you going to delete revisions every single time? That's tedious. What if you forget to delete them?

Thus, sharing .env files is **the exact opposite of the best practice of keeping configuration values secret**.

## The Problem of Different Configuration Values per Environment

At this point, engineers were already struggling, but an even more complex problem emerged.

It's the issue of **different configuration values for different environments**.

Particularly in web system development, there are various development environments, but generally they're divided as follows:

- **Local development environment**: The environment where developers work on their PCs/Macs
- **Staging environment**: An environment that reproduces what was developed in the same state as production (on the cloud)
- **Production environment**: Also called the live environment. The environment actually serving business or solutions

Depending on the development site, this may further split into multiple development environments, but this is generally the common structure.

However, these environments are all built as separate environments. Otherwise, it would be meaningless.

For example, the staging environment exists to test systems built by multiple developers in a production-like situation, but it must be separate from the production environment.

Otherwise, if a bug occurs in production, it could negatively impact the production environment that's running as a business or solution.

Therefore, connection information for servers and databases differs for each environment.

Using the earlier example, locally it would be:

```txt
DATABASE_URL=postgres://user:password@localhost:5432
```

But for staging:

```
DATABASE_URL=postgres://user:password@staging.example.com:5432
```

And a production version is also needed.

Thus, the environment variable name (key) is the same, but the value changes depending on the environment.

`dotenv` takes the following approach to this:

- `.env.local`
- `.env.staging`
- `.env.production`

Instead of just `.env`, you add a `.` followed by the environment name.

The environment names above are general examples—projects may also have `.env.develop` or `.env.test`.

However, what's common is that `.env.your_environment_name` should not be included in version control.

So how should you write `.gitignore`? Like this:

```
.env
.env.*
```

The asterisk (*) is a wildcard, meaning it matches any string after `.env.`.
That is, files named `.env` and `.env.(any string here)` are excluded from Git.

## Too Much Sensitive Information in Modern Times

Now, we've learned many things so far:

- **Keep configuration values secret with environment variables**
- **Use dotenv and .env files to manage environment variables in files**
- **However, sharing via email, chat, or file sharing services is dangerous**
- **.env files should not be included in version control systems like Git**
- **Create multiple .env files for each environment, like .env.staging**

Even more challenges emerge.

The problem of **too many configuration values**.

In the earlier example, we only defined one database URL, but in actual development environments, you handle many more environment variables.

It's common to work with multiple servers in parallel, and you might add faster key-value databases for caching.

Each of these has URLs, usernames, and passwords.

Configuration values quickly grow to 10, 20, and more.

It becomes hard to know what each configuration value name means.

In .env, if you prefix a line with `#`, it automatically becomes a comment.

However, the challenge of sharing .env files themselves remains.

This is where the `.env.example` approach was conceived.

# Writing `.env.example`

For example, let's say you write `.env` like this:

```text
# Main database URL. Using Postgres
DATABASE_URL=postgres://user:password@localhost:5432
```

This should work for the developer's local environment.

Then, prepare a file called `.env.example`.

The contents are as follows:

```text
# Copy .env.example, rename it to .env, then replace the values before use
# Main database URL. Using Postgres.
DATABASE_URL=postgres://sample_user_name:sample_password@example.com:5432
```

How's that? It's clearly a sample, and the comments convey that you should modify and use it.

And importantly, **`.env.example` can be managed in version control systems like Git**.

Since "example" means it's a sample, the values written there must only be **samples**.

However, the variable names (key names) are the same as `.env`, so developers can see what environment variable names are available.

If you're preparing .env files for each environment, it would look like:

- .env.local ← Not version controlled
- .env.local.example ← **Version controlled**
- .env.staging ← Not version controlled
- .env.staging.example ← **Version controlled**
- .env.production ← Not version controlled
- .env.production.example ← **Version controlled**

It's confusing, but this is a very important point.

So how do you version control only .example files?

In .gitignore, you write:

```text
.env
.env.*
!.env.example
!.env.*.example
```

This is how it looks. In .gitignore, when a line starts with `!`, that line's specification is negated. It's confusing, but it means "**ignore the item that ignores version control = version control it**".

## This .env Chaos Must Be Fixed

The situation up to this point, based on my personal experience, already existed around the mid-2010s.

But even though we've efficiently built up the logic—environment variables → .env → environment-specific .env → the invention of .example → using .gitignore—we still haven't fundamentally solved the challenge of **securely transmitting the environment variables themselves**.

Conversely, this challenge has been addressed from the mid-2010s to now through ad-hoc solutions by individual teams and developers, each with their own style.

In other words, it's been improvised, and there's no best practice that can be called **the standard**.

Of course, there are approaches to solve this:

- [AWS SSM](https://aws.amazon.com/systems-manager/)
- [Google Secret Manager](https://cloud.google.com/security/products/secret-manager)

These exist.

However, these tools have a strong flavor of being for managing environment variables to run production environments on AWS or GCP.

Also, since AWS and GCP have extensive services, managing account permissions requires careful attention.

In other words, they tend to be services where **managing the management** becomes a relatively large burden. (Clearly excessive just for .env management)

## Revisiting .env File Transmission Methods

I can't speak to what the current industry standard is (since I mostly work on in-house services), but my image of .env files was that when you join a project, they're sent to you via chat from your supervisor or senior colleagues.

This might still be the case somewhere.

In my experience, when I receive information via chat, I delete it after receiving it, but I don't really know if it's truly deleted.

For example, with chat tools like Slack or Chatwork, when you press the "delete" button to delete this information, it shows something like "*This message was deleted*".

Engineers who know better think, "Isn't that a soft delete?"

Even if it truly is a hard delete, you're temporarily storing sensitive information on Slack or Chatwork's servers.

If your development NDA stipulates that confidential information cannot be placed on third-party servers, this is a serious problem.

Send it via email? Out of the question. Email can be intercepted, so you should never do this.

Pass it on an encrypted USB drive? That's quite a retro method, but it's difficult in the age of remote work. Actually, it's a bit of a hassle even on the same floor. Many offices prohibit USB drives anyway.

The most secure method is to write it on a sticky note in the same office, hand it over, write it into .env, and then shred the note.

Hmm, is this like the 20th century?

So, as you can see, methods for securely sending environment variables are all still risky.

# Easier Transmission and Management

I apologize for not having a decisive solution for transmission, but some of you reading this far may have noticed:

"This is **not just a transmission problem, but management is super difficult** too, right?"

Yes. **Exactly!**

Different files for each environment, and there are multiple of them.

Also, environment variables tend to increase in number as development progresses.

Also, environment variable values sometimes change for various reasons.

One day you `git pull` and there's a new variable added to `.env.example`, and you have to ask someone for the new value.

This situation occurs.

It's not just transmission. When updated, you have to resend the new information repeatedly or rewrite it.

AWS SSM and GCP Secret Manager do consider these aspects, but as mentioned, they're somewhat excessive services.

# We Must Manage Environment Variables in the Cloud (But Safely)

If you've read this far carefully, you should understand how important configuration values are and how they should be kept secret.

On the other hand, there's the current challenge of having to proceed with a system like .env, which can be managed locally but has many tedious issues.

So our idea is that .env should be managed not just locally, but in the cloud (remotely).

Since both AWS and GCP effectively manage in the cloud, managing environment variables ≠ cloud is NG.

One important thing is that **managing environment variables in plaintext in the cloud is NG**.

Of course, it would be nice if we could manage locally only, but that has already reached its limits.

So the question becomes how to **safely manage environment variables in the cloud**.

The answer is **strong encryption**.

Actually, some chat and file sharing services already implement E2E encryption.
If so, sending → hard delete (if it truly is a hard delete) might be okay.

However, this doesn't fulfill the goal of **conveniently managing environment variables**.

Should you send to the team chat every time the .env.example file changes, and have the leader manually delete it a few hours later?

Should development members manually copy and paste from chat and overwrite .env?

In this day and age, that's a bit nonsensical.

So, in developing `bwsf`, we created the following concept.

## `bwsf` Development Concept

1. Environment variables are managed in the cloud (remotely) (however, with strong encryption)
2. Environment variables can be applied to local environments via CLI
3. Environment variables can be sent to the cloud (remotely) via CLI
4. .example files are excluded and should be Git-managed

These are concepts, but we considered detailed tech stacks for implementation.

Especially, what's the best approach for storing environment variables in the "cloud"? We thought about various options.

First, a managed SaaS was an initial strong option that came to mind.

However, important considerations quickly emerged that forced us to exclude it:

- What happens if that service ends?
- Is the encryption strength sufficient?
- There's a possibility of being affected by specification changes

On the other hand, self-developing a secure backend service also takes considerable effort.

Also, bwsf's purpose is to simplify .env management, not to create open-source secure storage.

So we decided to use something existing for the backend—specifically something that has **some flexibility in how data is structured** and has **high encryption strength**.

What was ultimately adopted is [Bitwarden](https://bitwarden.com/).

Bitwarden is a password manager that is also available as [open source](https://bitwarden.com/open-source/).

Regarding encryption, it uses **AES-256** with **600,000 PBKDF2 iterations**, which we considered sufficient for our requirements.

Next, functionality.

Basically, Bitwarden is a password manager, so it's designed to easily store login items (ID, password, TOTP), etc.

On the other hand, it can also store information other than passwords (such as credit card information or ID images).

What we focused on was the "Secure Note" format.

This is stored in a very simple blog-like format with a title and body text.

We decided to store the "project name" in the title and the .env file contents in JSON format in the "note body".

Also, Bitwarden has a concept of "organizations", and you can grant access permissions to users associated with that "organization".

This means you can allow only specific developers to access only specific project bwsf/Bitwarden note items.

If a member leaves the development team, you just need to remove them on Bitwarden.

Also, as mentioned earlier, **Bitwarden is open source**.

`bwsf` itself can be used with the cloud version of Bitwarden, but of course it can also be used with the open-source version.

This is very effective when there's a requirement not to place confidential information on third-party servers.

You can host Bitwarden on your own on-premises server or VPS and manage it there.

This is a requirement that even AWS SSM or Google Secret Manager cannot fulfill.

The existence of the `bw` command was also significant.

The `bw` command is Bitwarden's CLI, and while not all features are available, you can perform typical Bitwarden operations from the command line.

`bwsf` currently depends on the `bw` command for features like login.

However, thanks to this, we were able to develop `bwsf` quickly.

## In Closing

`bwsf` is still in development. There are many features we still want to add.

We extend our deep gratitude to Bitwarden and the Bitwarden open-source community for allowing us to develop with almost no backend development needed.

Also, if you see potential in `bwsf` and would like to help with development, you're always welcome. Please read the [Contributing Guidelines](https://github.com/b4m-oss/bwsf/blob/main/CONTRIBUTING.md) and join us.

-----

Kohki SHIKATA, CEO of Bicycle for Mind LLC
Dec 1st 2025

