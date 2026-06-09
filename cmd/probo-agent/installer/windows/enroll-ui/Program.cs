// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

using System.Text.Json;

namespace Probo.Agent.EnrollUI;

internal static class Program
{
    [STAThread]
    private static int Main()
    {
        Application.SetHighDpiMode(HighDpiMode.PerMonitorV2);
        Application.EnableVisualStyles();
        Application.SetCompatibleTextRenderingDefault(false);
        ApplicationConfiguration.Initialize();

        RegionsManifest manifest;
        try
        {
            manifest = RegionsManifest.Load();
        }
        catch (Exception ex)
        {
            MessageBox.Show(
                ex.Message,
                "Probo Device Posture Agent",
                MessageBoxButtons.OK,
                MessageBoxIcon.Error);
            return 2;
        }

        using var form = new EnrollmentForm(manifest);
        var result = form.ShowDialog();
        if (result != DialogResult.OK || form.Payload is null)
        {
            return 1;
        }

        Console.Out.WriteLine(JsonSerializer.Serialize(form.Payload));
        Console.Out.Flush();
        return 0;
    }
}
