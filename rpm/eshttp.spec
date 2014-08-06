Name:   eshttp
Version:        1.0
Release:        el6
Summary:        A distributed HTTP service for bulked Elasticseatch and AWS S3 indexing

Group:          Applications/Server
License:        Apache Licence 2.0
URL:            https://github.com/fangli/eshttp
SOURCE0:        https://github.com/fangli/eshttp/releases/download/1.0/eshttp.tar.gz
BuildRoot:      %(mktemp -ud %{_tmppath}/%{name}-%{version}-%{release}-XXXXXX)

%description
A distributed HTTP service for bulked Elasticseatch and AWS S3 indexing

%prep
%setup -c

%pre

%install
mv eshttp %{buildroot}%{_bindir}/
install -d %{buildroot}%{_initddir}
mv etc/init.d/eshttp %{buildroot}%{_initddir}/
install -d %{buildroot}%{_sysconfdir}/
mv etc/eshttp.conf %{buildroot}%{_sysconfdir}/eshttp.conf


%clean
rm -rf %{buildroot}

%files
%defattr(-,root,root,-)
%{_bindir}/*
%{_initddir}/*
%config %{_sysconfdir}/eshttp.conf

%post
/sbin/chkconfig --add eshttp

%preun
  if [ $1 = 0 ]; then
      # package is being erased, not upgraded
      /sbin/service eshttp stop
      /sbin/chkconfig --del eshttp
  fi

%postun
  if [ $1 = 0 ]; then
      echo "Uninstalling eshttp"
      rm %{_bindir}/eshttp
      # package is being erased
      # Any needed actions here on uninstalls
  else
      # Upgrade
      echo "Stopped, do not forget to run eshttp again"
  fi


%changelog
* Wed Aug 06 2014  Fang.li <fang.li@funplus.com>
  - Initial release
