// ============================================================================
// Renderizadores especializados
// ============================================================================
function RenderAnalise({ obj }: { obj: ObjetoTipo }) {
  if (!obj.questoes?.length) return null;
  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Fundamentação Jurídica
      </Typography>
      {obj.questoes.map((q, i) => (
        <Box key={i} mb={2}>
          <Typography variant="subtitle1" sx={{ fontWeight: "bold" }}>
            {q.tema ?? "Questão"}:
          </Typography>
          {q.paragrafos?.map((p, j) => (
            <Typography key={j} variant="body2" paragraph>
              {p}
            </Typography>
          ))}
        </Box>
      ))}
    </Box>
  );
}

function RenderSentenca({ obj }: { obj: ObjetoTipo }) {
  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Relatório
      </Typography>
      {obj.questoes
        ?.filter((q) => q.tipo === "relatorio")
        ?.flatMap((q) => q.paragrafos ?? [])
        ?.map((p, i) => (
          <Typography key={i} variant="body2" paragraph>
            {p}
          </Typography>
        ))}

      <Divider sx={{ my: 2 }} />

      <Typography variant="h6" gutterBottom>
        Fundamentação
      </Typography>
      {obj.questoes
        ?.filter((q) => q.tipo === "mérito" || q.tipo === "fundamentacao")
        ?.flatMap((q) => q.paragrafos ?? [])
        ?.map((p, i) => (
          <Typography key={i} variant="body2" paragraph>
            {p}
          </Typography>
        ))}

      {obj.dispositivo?.paragrafos && (
        <>
          <Divider sx={{ my: 2 }} />
          <Typography variant="h6" gutterBottom>
            Dispositivo
          </Typography>
          {obj.dispositivo.paragrafos.map((p, i) => (
            <Typography key={i} variant="body1" paragraph>
              {p}
            </Typography>
          ))}
        </>
      )}
    </Box>
  );
}

function RenderPeticao({ obj }: { obj: ObjetoTipo }) {
  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Exposição dos Fatos
      </Typography>
      {obj.corpo?.map((p, i) => (
        <Typography key={i} variant="body2" paragraph>
          {p}
        </Typography>
      ))}

      {obj.pedidos?.length && (
        <>
          <Divider sx={{ my: 2 }} />
          <Typography variant="h6" gutterBottom>
            Pedidos
          </Typography>
          {obj.pedidos.map((p, i) => (
            <Typography key={i} variant="body2" paragraph>
              {p}
            </Typography>
          ))}
        </>
      )}
    </Box>
  );
}

// fallback genérico
function RenderGenerico({ obj }: { obj: ObjetoTipo }) {
  return (
    <Box>
      <Typography variant="body2" color="text.secondary">
        Documento genérico ou estrutura não reconhecida.
      </Typography>
      <pre style={{ whiteSpace: "pre-wrap" }}>
        {JSON.stringify(obj, null, 2)}
      </pre>
    </Box>
  );
}
