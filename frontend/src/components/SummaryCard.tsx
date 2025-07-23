export default function SummaryCard({summary}:{summary?:string}) {
  if (!summary) return null;
  return (
    <div className="bg-gray-50 p-4 rounded-md border">
      <h4 className="font-semibold mb-1">AI Summary</h4>
      <p>{summary}</p>
    </div>
  );
}
