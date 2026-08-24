import {Locale, useTranslations} from 'next-intl';
import LocaleSwitcher from '@/components/LocaleSwitcher';
import {setRequestLocale} from 'next-intl/server';

export default function Some(props:any) {

    setRequestLocale(props.locale);
    const t = useTranslations('HomePage');

    return (<div>
        <LocaleSwitcher></LocaleSwitcher>
        <div>{t('title')} 333</div>
    </div>)
}
