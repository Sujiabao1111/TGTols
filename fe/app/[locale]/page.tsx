import {Locale, useTranslations} from 'next-intl';
import LocaleSwitcher from '@/components/LocaleSwitcher';
import {setRequestLocale} from 'next-intl/server';
import {use} from 'react';
export default function SomeHome({params}: PageProps<'/[locale]'>) {
    const {locale} = use(params);
     // Enable static rendering
    setRequestLocale(locale as Locale);
    const t = useTranslations('HomePage');
    return (<div>
        <LocaleSwitcher></LocaleSwitcher>
        <div>{t('title')} 233</div>
    </div>)
}
