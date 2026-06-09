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

using System.Diagnostics;

namespace Probo.Agent.EnrollUI;

internal sealed class EnrollmentForm : Form
{
    private static readonly Color AccentColor = Color.FromArgb(0, 122, 255);

    private readonly RegionsManifest _manifest;
    private readonly Dictionary<WorkspaceRegion, RadioButton> _regionRadios = new();
    private readonly TextBox _customHostField = new();
    private readonly Panel _employeeLinkPanel = new();
    private readonly LinkLabel _employeeLink = new();
    private readonly Label _employeeHint = new();
    private readonly TextBox _tokenField = new();
    private readonly Label _errorLabel = new();
    private readonly Button _enrollButton = new();

    private WorkspaceRegion _region = WorkspaceRegion.Us;

    internal EnrollmentPayload? Payload { get; private set; }

    internal EnrollmentForm(RegionsManifest manifest)
    {
        _manifest = manifest;

        Text = "Probo Device Posture Agent";
        FormBorderStyle = FormBorderStyle.FixedDialog;
        MaximizeBox = false;
        MinimizeBox = false;
        StartPosition = FormStartPosition.CenterScreen;
        ClientSize = new Size(500, 520);
        Font = new Font("Segoe UI", 9F, FontStyle.Regular, GraphicsUnit.Point);
        BackColor = SystemColors.Window;

        _errorLabel.AutoSize = true;
        _errorLabel.ForeColor = Color.Red;
        _errorLabel.MaximumSize = new Size(452, 0);
        _errorLabel.Margin = new Padding(0, 0, 0, 8);

        var layout = new TableLayoutPanel
        {
            Dock = DockStyle.Fill,
            ColumnCount = 1,
            RowCount = 5,
            Padding = new Padding(24),
        };
        layout.ColumnStyles.Add(new ColumnStyle(SizeType.Percent, 100F));
        layout.RowStyles.Add(new RowStyle(SizeType.AutoSize));
        layout.RowStyles.Add(new RowStyle(SizeType.AutoSize));
        layout.RowStyles.Add(new RowStyle(SizeType.AutoSize));
        layout.RowStyles.Add(new RowStyle(SizeType.AutoSize));
        layout.RowStyles.Add(new RowStyle(SizeType.Percent, 100F));

        layout.Controls.Add(BuildHeader(), 0, 0);
        layout.Controls.Add(BuildWorkspaceSection(), 0, 1);
        layout.Controls.Add(BuildTokenSection(), 0, 2);
        layout.Controls.Add(_errorLabel, 0, 3);
        layout.Controls.Add(BuildActions(), 0, 4);

        Controls.Add(layout);

        AcceptButton = _enrollButton;
        _regionRadios[WorkspaceRegion.Us].Checked = true;

        UpdateRegionUI();
    }

    private Control BuildHeader()
    {
        var panel = new FlowLayoutPanel
        {
            AutoSize = true,
            FlowDirection = FlowDirection.TopDown,
            WrapContents = false,
            Dock = DockStyle.Top,
            Margin = new Padding(0, 0, 0, 12),
        };

        panel.Controls.Add(new Label
        {
            Text = "Connect this PC to Probo",
            Font = new Font(Font.FontFamily, 16F, FontStyle.Bold),
            AutoSize = true,
            Margin = new Padding(0, 0, 0, 6),
        });

        panel.Controls.Add(new Label
        {
            Text =
                "Choose where your workspace is hosted, open your employee "
                + "page to copy an enrollment token, then paste it below.",
            AutoSize = true,
            MaximumSize = new Size(452, 0),
            ForeColor = SystemColors.GrayText,
            Margin = new Padding(0),
        });

        return panel;
    }

    private Control BuildWorkspaceSection()
    {
        var group = new GroupBox
        {
            Text = "Workspace",
            AutoSize = true,
            Dock = DockStyle.Top,
            Margin = new Padding(0, 0, 0, 12),
            Padding = new Padding(10, 20, 10, 10),
            Width = 452,
        };

        var panel = new FlowLayoutPanel
        {
            AutoSize = true,
            FlowDirection = FlowDirection.TopDown,
            WrapContents = false,
            Dock = DockStyle.Fill,
        };

        var cards = new TableLayoutPanel
        {
            AutoSize = true,
            ColumnCount = 3,
            RowCount = 1,
            Margin = new Padding(0, 0, 0, 8),
        };
        cards.ColumnStyles.Add(new ColumnStyle(SizeType.Percent, 33.33F));
        cards.ColumnStyles.Add(new ColumnStyle(SizeType.Percent, 33.33F));
        cards.ColumnStyles.Add(new ColumnStyle(SizeType.Percent, 33.34F));

        foreach (var region in new[] { WorkspaceRegion.Us, WorkspaceRegion.Eu, WorkspaceRegion.SelfHosted })
        {
            var radio = CreateRegionRadio(region);
            _regionRadios[region] = radio;
            cards.Controls.Add(radio);
        }

        panel.Controls.Add(cards);

        _customHostField.PlaceholderText = "probo.example.com";
        _customHostField.Width = 420;
        _customHostField.Margin = new Padding(0, 0, 0, 8);
        _customHostField.TextChanged += (_, _) => UpdateEmployeeLink();
        panel.Controls.Add(_customHostField);

        _employeeLinkPanel.AutoSize = true;
        _employeeLinkPanel.Padding = new Padding(10);
        _employeeLinkPanel.BackColor = SystemColors.Control;
        _employeeLinkPanel.Margin = new Padding(0);
        _employeeLinkPanel.Width = 420;

        _employeeHint.AutoSize = true;
        _employeeHint.MaximumSize = new Size(400, 0);
        _employeeHint.ForeColor = SystemColors.GrayText;
        _employeeHint.Text = _manifest.EmployeePageHint;
        _employeeHint.Margin = new Padding(0, 0, 0, 6);

        _employeeLink.AutoSize = true;
        _employeeLink.Text = "Open employee page";
        _employeeLink.LinkColor = AccentColor;
        _employeeLink.ActiveLinkColor = AccentColor;
        _employeeLink.VisitedLinkColor = AccentColor;
        _employeeLink.LinkBehavior = LinkBehavior.HoverUnderline;
        _employeeLink.Click += (_, _) => OpenEmployeePage();

        _employeeLinkPanel.Controls.Add(_employeeHint);
        _employeeLinkPanel.Controls.Add(_employeeLink);
        panel.Controls.Add(_employeeLinkPanel);

        group.Controls.Add(panel);
        return group;
    }

    private RadioButton CreateRegionRadio(WorkspaceRegion region)
    {
        var entry = _manifest.Region(region);
        var radio = new RadioButton
        {
            Appearance = Appearance.Button,
            AutoSize = false,
            Size = new Size(130, 110),
            Text = $"{entry.Flag}{Environment.NewLine}{entry.Title}{Environment.NewLine}{entry.Subtitle}",
            TextAlign = ContentAlignment.MiddleCenter,
            FlatStyle = FlatStyle.Flat,
            Margin = new Padding(0, 0, 8, 0),
            TabStop = true,
            Tag = region,
        };

        radio.FlatAppearance.BorderColor = SystemColors.ControlDark;
        radio.FlatAppearance.CheckedBackColor = Color.FromArgb(26, 122, 255);
        radio.FlatAppearance.BorderSize = 1;

        radio.CheckedChanged += (_, _) =>
        {
            if (radio.Checked)
            {
                SelectRegion(region);
            }
        };

        return radio;
    }

    private Control BuildTokenSection()
    {
        var panel = new FlowLayoutPanel
        {
            AutoSize = true,
            FlowDirection = FlowDirection.TopDown,
            WrapContents = false,
            Margin = new Padding(0, 0, 0, 8),
        };

        panel.Controls.Add(new Label
        {
            Text = "Enrollment token",
            Font = new Font(Font, FontStyle.Bold),
            AutoSize = true,
            Margin = new Padding(0, 0, 0, 8),
        });

        _tokenField.Width = 452;
        _tokenField.UseSystemPasswordChar = true;
        _tokenField.PlaceholderText = "Paste token here";
        panel.Controls.Add(_tokenField);

        return panel;
    }

    private Control BuildActions()
    {
        var panel = new Panel
        {
            Dock = DockStyle.Bottom,
            Height = 36,
        };

        var cancel = new Button
        {
            Text = "Cancel",
            DialogResult = DialogResult.Cancel,
            AutoSize = true,
            Location = new Point(0, 4),
        };
        CancelButton = cancel;

        _enrollButton.Text = "Enroll device";
        _enrollButton.AutoSize = true;
        _enrollButton.BackColor = AccentColor;
        _enrollButton.ForeColor = Color.White;
        _enrollButton.FlatStyle = FlatStyle.Flat;
        _enrollButton.FlatAppearance.BorderSize = 0;
        _enrollButton.Padding = new Padding(12, 4, 12, 4);
        _enrollButton.Click += (_, _) => Submit();

        panel.Controls.Add(cancel);
        panel.Controls.Add(_enrollButton);

        panel.Resize += (_, _) =>
        {
            _enrollButton.Location = new Point(panel.ClientSize.Width - _enrollButton.Width, 0);
        };
        _enrollButton.Location = new Point(panel.ClientSize.Width - _enrollButton.Width, 0);

        return panel;
    }

    private void SelectRegion(WorkspaceRegion region)
    {
        _region = region;
        UpdateRegionUI();
    }

    private void UpdateRegionUI()
    {
        _customHostField.Visible = _region == WorkspaceRegion.SelfHosted;
        UpdateEmployeeLink();
    }

    private void UpdateEmployeeLink()
    {
        var url = _region.ConsoleBaseURL(_manifest, _customHostField.Text);
        if (url is not null)
        {
            _employeeHint.Text = _manifest.EmployeePageHint;
            _employeeLinkPanel.Visible = true;
            _employeeLink.Tag = url;
            _employeeLink.Enabled = true;
            return;
        }

        if (_region == WorkspaceRegion.SelfHosted)
        {
            _employeeHint.Text = _manifest.SelfHostedHostnameHint;
            _employeeLinkPanel.Visible = true;
            _employeeLink.Tag = null;
            _employeeLink.Enabled = false;
            return;
        }

        _employeeLinkPanel.Visible = false;
    }

    private void OpenEmployeePage()
    {
        if (_employeeLink.Tag is not string url)
        {
            return;
        }

        try
        {
            Process.Start(new ProcessStartInfo
            {
                FileName = url,
                UseShellExecute = true,
            });
        }
        catch (Exception ex)
        {
            _errorLabel.Text = ex.Message;
        }
    }

    private void Submit()
    {
        _errorLabel.Text = "";

        var serverURL = _region.ResolveServerURL(
            _manifest,
            _customHostField.Text,
            out var errorMessage);
        if (serverURL is null)
        {
            _errorLabel.Text = errorMessage ?? "Invalid workspace.";
            return;
        }

        var token = _tokenField.Text.Trim();
        if (token.Length == 0)
        {
            _errorLabel.Text = "Enrollment token is required.";
            return;
        }

        Payload = new EnrollmentPayload
        {
            ServerURL = serverURL,
            EnrollmentToken = token,
        };

        DialogResult = DialogResult.OK;
        Close();
    }
}
