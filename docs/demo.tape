# Records docs/demo.gif, the README's demo, with vhs
# (https://github.com/charmbracelet/vhs). Re-record after changes to the
# table output so the README never shows stale output.
#
# With vhs installed (it needs ttyd and ffmpeg), and a shorstop on PATH:
#
#   make build && PATH="$PWD/bin:$PATH" vhs docs/demo.tape
#
# Or in Docker, with a Linux release binary (copied in, so no file sharing
# needs configuring). The vhs image lacks git and jq, and its Debian base
# needs --allow-releaseinfo-change for apt:
#
#   printf 'FROM ghcr.io/charmbracelet/vhs:v0.10.0\nRUN apt-get update -qq --allow-releaseinfo-change && apt-get install -y -qq --no-install-recommends git jq ca-certificates\n' | docker build -t shorstop-vhs -
#   c=$(docker create -w /vhs shorstop-vhs docs/demo.tape)
#   docker cp docs "$c:/vhs/" && docker cp shorstop "$c:/usr/local/bin/shorstop"
#   docker start -a "$c" && docker cp "$c:/vhs/docs/demo.gif" docs/ && docker rm "$c"
#
# It scans the public demo repository (shorstophq/shorstop-demo), so the
# recording shows a real project, not a fixture made to look good.

Output docs/demo.gif

Set Shell "bash"
Set FontSize 14
Set Width 1500
Set Height 760
Set Padding 20
Set Theme "Dracula"
Set TypingSpeed 45ms

Hide
Type `export PS1='$ ' && cd "$(mktemp -d)" && git clone -q https://github.com/shorstophq/shorstop-demo . && clear`
Enter
Sleep 4s
Show

Type "# A small service with RSA, ECDSA, MD5, TLS 1.0 and a 3DES cipher string"
Sleep 600ms
Enter
Type "ls -R"
Enter
Sleep 1.5s

Hide
Type "clear"
Enter
Show
Type "shorstop scan | head -n 29"
Sleep 400ms
Enter
Sleep 7s

Type "# In CI: fail the build on critical findings"
Sleep 400ms
Enter
Type "shorstop scan --fail-on critical > /dev/null; echo exit $?"
Sleep 400ms
Enter
Sleep 2.5s

Type "# A CycloneDX 1.6 CBOM for your crypto inventory"
Sleep 400ms
Enter
Type "shorstop scan -f cbom -o cbom.json && jq '.components | length' cbom.json"
Sleep 400ms
Enter
Sleep 3.5s
