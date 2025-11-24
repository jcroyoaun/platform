export default function Footer() {
  return (
    <footer style={{ width: '100%', background: '#0f172a', color: '#94a3b8', padding: '3rem 2rem', marginTop: '3rem', boxSizing: 'border-box', borderTop: '3px solid #10b981' }}>
      <div style={{ maxWidth: '1200px', margin: '0 auto', width: '100%', boxSizing: 'border-box' }}>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '2rem', marginBottom: '2rem' }}>
          <div>
            <h3 style={{ color: 'white', fontWeight: 700, fontSize: '1.125rem', margin: '0 0 0.5rem 0' }}>💰 TotalComp MX</h3>
            <p style={{ margin: 0, fontSize: '0.875rem', lineHeight: 1.6 }}>
              Comparador de compensación total para México.<br />
              Actualizada con las tablas ISR 2025, UMA ($3,439.46) y SMG ($278.80).
            </p>
          </div>
          <div style={{ textAlign: 'right', fontSize: '0.75rem' }}>
            <p style={{ margin: 0 }}>Desarrollado con Go y HTML5.</p>
            <p style={{ margin: '0.5rem 0 0 0', color: '#64748b' }}>Compara ofertas laborales • Sueldos y Salarios vs RESICO</p>
          </div>
        </div>

        <div style={{ borderTop: '1px solid #1e293b', paddingTop: '1.5rem', fontSize: '0.75rem', textAlign: 'justify', color: '#64748b', lineHeight: 1.7 }}>
          <p style={{ margin: '0 0 0.75rem 0' }}>
            <strong style={{ color: '#fbbf24' }}>⚠️ Aviso Legal:</strong> Esta herramienta es un simulador de uso exclusivamente informativo y didáctico.
            Los cálculos presentados son estimaciones basadas en la legislación fiscal vigente (LISR, LSS, RMF 2025)
            pero <strong style={{ color: '#94a3b8' }}>no constituyen asesoría fiscal, contable o legal profesional</strong>.
          </p>
          <p style={{ margin: '0 0 0.75rem 0' }}>
            El desarrollador no se hace responsable por discrepancias con los cálculos oficiales del SAT o el IMSS,
            ni por la toma de decisiones financieras basadas en estos resultados. Se recomienda consultar a un
            contador público certificado para determinaciones fiscales definitivas.
          </p>
          <p style={{ margin: 0 }}>
            <strong style={{ color: '#10b981' }}>🔒 Privacidad:</strong> No recabamos, almacenamos ni compartimos datos personales.
            Todos los cálculos se realizan de forma anónima y los datos ingresados se eliminan al cerrar la sesión.
          </p>
        </div>

        <div style={{ textAlign: 'center', marginTop: '1.5rem', paddingTop: '1rem', borderTop: '1px solid #1e293b' }}>
          <div style={{ display: 'flex', justifyContent: 'center', gap: '1.5rem', flexWrap: 'wrap', marginBottom: '0.75rem' }}>
            <a
              href="/privacy"
              style={{ color: '#10b981', textDecoration: 'none', fontSize: '0.8rem', fontWeight: 500, transition: 'color 0.2s' }}
              onMouseOver={(e) => (e.currentTarget.style.color = '#059669')}
              onMouseOut={(e) => (e.currentTarget.style.color = '#10b981')}
            >
              🔒 Aviso de Privacidad
            </a>
            <span style={{ color: '#334155' }}>|</span>
            <a
              href="/terms"
              style={{ color: '#10b981', textDecoration: 'none', fontSize: '0.8rem', fontWeight: 500, transition: 'color 0.2s' }}
              onMouseOver={(e) => (e.currentTarget.style.color = '#059669')}
              onMouseOut={(e) => (e.currentTarget.style.color = '#10b981')}
            >
              📋 Términos y Condiciones
            </a>
          </div>
          <div style={{ fontSize: '0.7rem', color: '#475569' }}>
            &copy; 2025 TotalComp MX. Todos los derechos reservados. • Versión 2025 (Beta)
          </div>
        </div>
      </div>
    </footer>
  )
}
