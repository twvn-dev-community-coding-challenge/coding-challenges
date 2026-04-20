import { screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';

import { renderWithTheme } from '../../test-utils';
import { MembershipSmsForm } from './membership-sms-form';

const baseValues = {
  messageId: 'msg-1',
  recipient: 'user@example.com',
  countryCode: 'VN',
  phoneNumber: '',
  content: 'Hello',
} as const;

describe('MembershipSmsForm', () => {
  it('renders all form fields: Message ID, Recipient, Country Code (select), Phone Number, Message Content', () => {
    renderWithTheme(
      <MembershipSmsForm
        values={baseValues}
        onChange={vi.fn()}
        onSubmit={vi.fn()}
      />,
    );

    expect(screen.getByLabelText(/message id/i)).toBeTruthy();
    expect(screen.getByLabelText(/recipient/i)).toBeTruthy();
    expect(screen.getByLabelText(/country code/i)).toBeTruthy();
    expect(screen.getByLabelText(/phone number/i)).toBeTruthy();
    expect(screen.getByLabelText(/message content/i)).toBeTruthy();
  });

  it('each field has an accessible label', () => {
    renderWithTheme(
      <MembershipSmsForm
        values={baseValues}
        onChange={vi.fn()}
        onSubmit={vi.fn()}
      />,
    );

    expect(screen.getByLabelText(/message id/i)).toBeTruthy();
    expect(screen.getByLabelText(/recipient/i)).toBeTruthy();
    expect(screen.getByLabelText(/country code/i)).toBeTruthy();
    expect(screen.getByLabelText(/phone number/i)).toBeTruthy();
    expect(screen.getByLabelText(/message content/i)).toBeTruthy();
  });

  it('calls onChange when Message ID is edited', () => {
    const onChange = vi.fn();
    renderWithTheme(
      <MembershipSmsForm
        values={baseValues}
        onChange={onChange}
        onSubmit={vi.fn()}
      />,
    );

    fireEvent.change(screen.getByLabelText(/message id/i), {
      target: { value: 'new-id' },
    });

    expect(onChange).toHaveBeenCalledWith({ messageId: 'new-id' });
  });

  it('calls onChange when Recipient is edited', () => {
    const onChange = vi.fn();
    renderWithTheme(
      <MembershipSmsForm
        values={baseValues}
        onChange={onChange}
        onSubmit={vi.fn()}
      />,
    );

    fireEvent.change(screen.getByLabelText(/recipient/i), {
      target: { value: 'r@x.com' },
    });

    expect(onChange).toHaveBeenCalledWith({ recipient: 'r@x.com' });
  });

  it('calls onChange when Country Code is changed (select: VN +84, TH +66, SG +65, PH +63)', () => {
    const onChange = vi.fn();
    renderWithTheme(
      <MembershipSmsForm
        values={baseValues}
        onChange={onChange}
        onSubmit={vi.fn()}
      />,
    );

    const select = screen.getByLabelText(/country code/i) as HTMLSelectElement;
    fireEvent.change(select, { target: { value: 'TH' } });

    expect(onChange).toHaveBeenCalledWith({ countryCode: 'TH' });
  });

  it('calls onChange when Phone Number is edited', () => {
    const onChange = vi.fn();
    renderWithTheme(
      <MembershipSmsForm
        values={baseValues}
        onChange={onChange}
        onSubmit={vi.fn()}
      />,
    );

    fireEvent.change(screen.getByLabelText(/phone number/i), {
      target: { value: '+84901' },
    });

    expect(onChange).toHaveBeenCalledWith({ phoneNumber: '+84901' });
  });

  it('calls onPhoneBlur when the phone field loses focus', () => {
    const onPhoneBlur = vi.fn();
    renderWithTheme(
      <MembershipSmsForm
        values={baseValues}
        onChange={vi.fn()}
        onSubmit={vi.fn()}
        onPhoneBlur={onPhoneBlur}
      />,
    );

    fireEvent.blur(screen.getByLabelText(/phone number/i));
    expect(onPhoneBlur).toHaveBeenCalledTimes(1);
  });

  it('shows field validation errors when fieldErrors are set', () => {
    renderWithTheme(
      <MembershipSmsForm
        values={{ ...baseValues, messageId: '', recipient: '', phoneNumber: '', content: '' }}
        onChange={vi.fn()}
        onSubmit={vi.fn()}
        fieldErrors={{
          messageId: 'Message ID is required.',
          recipient: 'Recipient is required.',
          phoneNumber: 'Phone number is required',
          content: 'Message content is required.',
        }}
      />,
    );

    expect(screen.getByText('Message ID is required.')).toBeTruthy();
    expect(screen.getByText('Recipient is required.')).toBeTruthy();
    expect(screen.getByText('Phone number is required')).toBeTruthy();
    expect(screen.getByText('Message content is required.')).toBeTruthy();

    expect((screen.getByLabelText(/message id/i) as HTMLInputElement).getAttribute('aria-invalid')).toBe('true');
    expect((screen.getByLabelText(/recipient/i) as HTMLInputElement).getAttribute('aria-invalid')).toBe('true');
    expect((screen.getByLabelText(/phone number/i) as HTMLInputElement).getAttribute('aria-invalid')).toBe('true');
    expect((screen.getByLabelText(/message content/i) as HTMLTextAreaElement).getAttribute('aria-invalid')).toBe('true');
  });

  it('calls onChange when Message Content is edited', () => {
    const onChange = vi.fn();
    renderWithTheme(
      <MembershipSmsForm
        values={baseValues}
        onChange={onChange}
        onSubmit={vi.fn()}
      />,
    );

    fireEvent.change(screen.getByLabelText(/message content/i), {
      target: { value: 'Updated body' },
    });

    expect(onChange).toHaveBeenCalledWith({ content: 'Updated body' });
  });

  it('all fields are disabled when disabled prop is true', () => {
    renderWithTheme(
      <MembershipSmsForm
        values={baseValues}
        onChange={vi.fn()}
        onSubmit={vi.fn()}
        disabled
      />,
    );

    expect((screen.getByLabelText(/message id/i) as HTMLInputElement).disabled).toBe(true);
    expect((screen.getByLabelText(/recipient/i) as HTMLInputElement).disabled).toBe(true);
    expect((screen.getByLabelText(/country code/i) as HTMLSelectElement).disabled).toBe(true);
    expect((screen.getByLabelText(/phone number/i) as HTMLInputElement).disabled).toBe(true);
    expect((screen.getByLabelText(/message content/i) as HTMLTextAreaElement).disabled).toBe(true);
  });

  it('renders a "Send SMS" button', () => {
    renderWithTheme(
      <MembershipSmsForm
        values={baseValues}
        onChange={vi.fn()}
        onSubmit={vi.fn()}
      />,
    );

    expect(screen.getByRole('button', { name: /send sms/i })).toBeTruthy();
  });

  it('calls onSubmit when form is submitted', () => {
    const onSubmit = vi.fn();
    renderWithTheme(
      <MembershipSmsForm
        values={baseValues}
        onChange={vi.fn()}
        onSubmit={onSubmit}
      />,
    );

    fireEvent.submit(screen.getByRole('button', { name: /send sms/i }).closest('form') as HTMLFormElement);

    expect(onSubmit).toHaveBeenCalledTimes(1);
  });

  it('Send SMS button is disabled when disabled prop is true', () => {
    renderWithTheme(
      <MembershipSmsForm
        values={baseValues}
        onChange={vi.fn()}
        onSubmit={vi.fn()}
        disabled
      />,
    );

    expect(
      (screen.getByRole('button', { name: /send sms/i }) as HTMLButtonElement).disabled,
    ).toBe(true);
  });
});
