// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { Button, Section, Text } from 'react-email';
import * as React from 'react';
import EmailLayout, { bodyText, button, buttonContainer, footerText } from './components/EmailLayout';
import { NotificationSettingsLink } from './components/NotificationSettingsLink';

const taskList = {
  ...bodyText,
  paddingLeft: '20px',
};

export const TaskMentioned = () => {
  return (
    <EmailLayout subject={'You were mentioned in tasks at {{.OrganizationName}}'}>
      <Text style={bodyText}>
        You're receiving this message because <strong>{'{{.OrganizationName}}'}</strong> mentioned you in these tasks:
      </Text>

      <ul style={taskList} dangerouslySetInnerHTML={{ __html: '{{range .Tasks}}<li><a href="{{.URL}}">{{.Title}}</a></li>{{end}}' }} />

      <Text style={bodyText}>
        Open your tasks to review them:
      </Text>

      <Section style={buttonContainer}>
        <Button style={button} href={'{{.TasksUrl}}'}>
          View tasks
        </Button>
      </Section>

      <NotificationSettingsLink />
      <Text style={footerText}>
        Best regards,
      </Text>
    </EmailLayout>
  );
};

export default TaskMentioned;
