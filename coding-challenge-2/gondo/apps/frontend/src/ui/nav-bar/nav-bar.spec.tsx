import { screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router';

import { renderWithTheme } from '../../test-utils';
import { NavBar } from './nav-bar';

const links = [
  { to: '/membership', label: 'Membership Registration' },
  { to: '/booking', label: 'Booking Notifications' },
];

describe('NavBar', () => {
  it('renders all nav links', () => {
    renderWithTheme(
      <MemoryRouter>
        <NavBar links={links} />
      </MemoryRouter>,
    );
    expect(screen.getByRole('link', { name: 'Membership Registration' })).toBeTruthy();
    expect(screen.getByRole('link', { name: 'Booking Notifications' })).toBeTruthy();
  });

  it('renders a navigation landmark', () => {
    renderWithTheme(
      <MemoryRouter>
        <NavBar links={links} />
      </MemoryRouter>,
    );
    expect(screen.getByRole('navigation')).toBeTruthy();
  });

  it('highlights active link based on current route', () => {
    renderWithTheme(
      <MemoryRouter initialEntries={['/membership']}>
        <NavBar links={links} />
      </MemoryRouter>,
    );
    const membershipLink = screen.getByRole('link', { name: 'Membership Registration' });
    expect(membershipLink.getAttribute('aria-current')).toBe('page');
  });

  it('renders app title/brand', () => {
    renderWithTheme(
      <MemoryRouter>
        <NavBar links={links} />
      </MemoryRouter>,
    );
    expect(screen.getByText('Gondo Platform')).toBeTruthy();
  });
});
