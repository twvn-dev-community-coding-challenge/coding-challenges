import SmsOutlined from '@mui/icons-material/SmsOutlined';
import { AppBar, Box, Toolbar, Typography } from '@mui/material';
import { NavLink } from 'react-router';

interface NavLinkItem {
  readonly to: string;
  readonly label: string;
}

interface NavBarProps {
  readonly links: readonly NavLinkItem[];
}

export const NavBar = ({ links }: NavBarProps) => {
  return (
    <AppBar
      component="nav"
      position="static"
      color="secondary"
      elevation={2}
      aria-label="Main"
    >
      <Toolbar
        sx={{
          alignItems: 'flex-start',
          flexWrap: 'wrap',
          gap: { xs: 1.5, sm: 2 },
          py: { xs: 1.5, sm: 1.25 },
          rowGap: 1.25,
        }}
      >
        <Box
          sx={{
            display: 'flex',
            alignItems: 'center',
            gap: 1.25,
            flexGrow: 1,
            minWidth: 0,
            mr: { xs: 0, md: 1 },
          }}
        >
          <SmsOutlined sx={{ fontSize: 28, opacity: 0.95 }} aria-hidden />
          <Typography variant="h6" component="span" noWrap>
            Gondo SMS Platform
          </Typography>
        </Box>
        <Box
          component="ul"
          sx={{
            display: 'flex',
            flexWrap: 'wrap',
            alignItems: 'center',
            justifyContent: 'flex-end',
            gap: 0.75,
            listStyle: 'none',
            m: 0,
            p: 0,
            flex: { xs: '1 1 100%', md: '0 1 auto' },
          }}
        >
          {links.map((link) => (
            <Box component="li" key={link.to}>
              <NavLink to={link.to} style={{ textDecoration: 'none' }}>
                {({ isActive }) => (
                  <Box
                    component="span"
                    sx={{
                      display: 'inline-flex',
                      alignItems: 'center',
                      px: 2,
                      py: 0.75,
                      borderRadius: 999,
                      color: 'common.white',
                      fontWeight: isActive ? 600 : 500,
                      fontSize: '0.9375rem',
                      lineHeight: 1.5,
                      bgcolor: isActive
                        ? 'rgba(255, 255, 255, 0.18)'
                        : 'transparent',
                      transition: (theme) =>
                        theme.transitions.create(
                          ['background-color', 'font-weight'],
                          { duration: theme.transitions.duration.shorter },
                        ),
                      '&:hover': {
                        bgcolor: isActive
                          ? 'rgba(255, 255, 255, 0.24)'
                          : 'rgba(255, 255, 255, 0.08)',
                      },
                    }}
                  >
                    {link.label}
                  </Box>
                )}
              </NavLink>
            </Box>
          ))}
        </Box>
      </Toolbar>
    </AppBar>
  );
};
