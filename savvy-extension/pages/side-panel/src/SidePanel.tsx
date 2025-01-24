import '@src/SidePanel.css';
import { withErrorBoundary, withSuspense } from '@extension/shared';
import { HistoryViewer } from './components/HistoryViewer';

const SidePanel = () => {
  return (
    <div className="w-full">
      <HistoryViewer />
    </div>
  );
};

export default withErrorBoundary(withSuspense(SidePanel, <div> Loading ... </div>), <div> Error Occur </div>);
