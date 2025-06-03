import { Link } from '@tanstack/react-router';
import { Fragment } from 'react/jsx-runtime';
import { ServiceServiceAdminDto } from '~/gen';

function ServiceAdministrators({ admins }: { admins: ServiceServiceAdminDto[] }) {
  if (admins.length === 0) {
    return null;
  }

  return (
    <>
      Service Administrators:{' '}
      {admins.map((admin, index) => (
        <Fragment key={admin.userId}>
          <Link className="link" to="/users/$userId" params={{ userId: admin.userId }}>
            {admin.userName}
          </Link>
          {index < admins.length - 1 && ', '}
        </Fragment>
      ))}
    </>
  );
}

export default ServiceAdministrators;
