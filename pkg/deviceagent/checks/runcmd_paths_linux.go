// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package checks

var linuxCommandPaths = map[string][]string{
	"firewall-cmd": {"/usr/bin/firewall-cmd"},
	"gsettings":    {"/usr/bin/gsettings"},
	"iptables":     {"/usr/sbin/iptables", "/sbin/iptables", "/usr/bin/iptables"},
	"lsblk":        {"/usr/bin/lsblk", "/bin/lsblk"},
	"nft":          {"/usr/sbin/nft", "/sbin/nft"},
	"systemctl":    {"/usr/bin/systemctl", "/bin/systemctl"},
	"timedatectl":  {"/usr/bin/timedatectl", "/bin/timedatectl"},
	"ufw":          {"/usr/sbin/ufw", "/sbin/ufw"},
}

func commandCandidates(cmd string) []string {
	return linuxCommandPaths[cmd]
}
