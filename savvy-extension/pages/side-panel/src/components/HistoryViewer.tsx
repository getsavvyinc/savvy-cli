import React, { useCallback, useEffect, useState } from 'react';
import { Button } from '@src/components/ui/button';
import { Checkbox } from '@src/components/ui/checkbox';
import { Toaster } from '@src/components/ui/sonner';
import { toast } from 'sonner';
import { Badge } from '@src/components/ui/badge';
import { ExternalLink, ChevronRight, ClipboardIcon } from 'lucide-react';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@src/components/ui/select';
import { useLocalClient } from '@extension/shared/lib/hooks/useAPI';
import { isAxiosError } from 'axios';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@src/components/ui/tooltip';
import { ScrollArea } from '@src/components/ui/scroll-area';

interface HistoryItem extends chrome.history.HistoryItem {
  isSelected?: boolean;
}

interface HistoryViewerProps {} // Empty interface since component takes no props

const TIME_RANGES = [
  { label: '30 mins', hours: 0.5 },
  { label: '1 hour', hours: 1 },
  { label: '2 hours', hours: 2 },
  { label: '3 hours', hours: 3 },
  { label: '5 hours', hours: 5 },
  { label: '8 hours', hours: 8 },
  { label: '12 hours', hours: 12 },
  { label: '16 hours', hours: 16 },
  { label: '20 hours', hours: 20 },
  { label: '24 hours', hours: 24 },
  { label: '48 hours', hours: 48 },
  { label: '72 hours', hours: 72 },
  { label: '96 hours', hours: 96 },
  { label: '120 hours', hours: 120 },
];

const DENIED_DOMAINS = [
  'chrome://',
  'x.com',
  'facebook.com',
  'linkedin.com',
  'instagram.com',
  'twitter.com',
  'reddit.com',
  'twitch.tv',
  'zoom.us',
  'zoom.com',
];

const ALLOWED_DOMAINS = [
  // Developer tools & documentation
  'getsavvy.so',
  'github', // Matches github.com, github.internal etc
  'stackoverflow',
  'gitlab',
  'bitbucket',

  // Error monitoring & logging
  'rollbar', // Matches rollbar.com, rollbar.internal etc
  'splunk',
  'datadog',
  'sentry',
  'bugsnag',
  'raygun',
  'signoz',
  'harness',
  'metabase',
  'mode',
  'posthog',
  'postman',
  'statsig',
  'replit',
  'repl.it',
  'launchdarkly',

  // Cloud platforms
  'aws.', // Matches aws.amazon.com, aws.internal etc
  'amazon',
  'console.aws',
  'cloud.google',
  'gcp', // Common abbreviation for Google Cloud Platform
  'azure',
  'render',
  'railway',
  'vercel',
  'northflank',
  'fly.io',
  'jam.dev',
  'netlify',

  // ai tools
  'julius',
  'openai.com',
  'claude.ai',
  'anthropic',
  'perplexity.ai',

  // Monitoring & APM
  'grafana',
  'newrelic',
  'oodle',
  'prometheus',
  'kibana',
  'elasticsearch',
  'honeycomb.io',
  'elk', // Common abbreviation for Elasticsearch, Logstash, Kibana

  // CI/CD
  'jenkins',
  'circleci',
  'travis',
  'teamcity',
  'ci.', // The dot after ci is important to avoid false positives

  // Support tools
  'intercom',
  'zendesk',
  'salesforce',
  'sfdc',
  'freshdesk',
  'pylon',
  'front',

  //incident & status page tools
  'incident',
  'opsgenie',
  'atlassian',
  'pagerduty',
  'statuspage',
  'status.',

  // Common internal domains
  'monitoring',
  'logs',
  'metrics',
  'debug',
  'trace',
  'apm',
  'observability',
  'ops',
  'devops',
];

const getHostname = (url: string) => {
  try {
    return new URL(url).hostname;
  } catch {
    return url;
  }
};

export const HistoryViewer: React.FC<HistoryViewerProps> = () => {
  const [selectedHours, setSelectedHours] = useState<number>(1);
  const [history, setHistory] = useState<HistoryItem[]>([]);
  const [loading, setLoading] = useState(false);
  const allowedDomains = ALLOWED_DOMAINS;
  const { client } = useLocalClient();

  const fetchHistory = useCallback(async () => {
    setLoading(true);
    try {
      const startTime = new Date(Date.now() - selectedHours * 60 * 60 * 1000).getTime();
      const items = await chrome.history.search({
        text: '',
        startTime,
        maxResults: 10000,
      });

      // filter out denied domains
      const filterOutDeniedDomains = items.filter(item => !DENIED_DOMAINS.some(domain => item.url!.includes(domain)));
      // Filter items by allowed domains
      const filteredItems = filterOutDeniedDomains.filter(item =>
        item.url ? allowedDomains.some(domain => item.url!.includes(domain)) : false,
      );

      setHistory(filteredItems);
    } catch (error) {
      console.error('Error fetching history:', error);
    } finally {
      setLoading(false);
    }
  }, [selectedHours, allowedDomains]);

  useEffect(() => {
    void fetchHistory();
  }, [fetchHistory]);

  const handleItemSelect = (index: number) => {
    setHistory(prevHistory =>
      prevHistory.map((item, i) => (i === index ? { ...item, isSelected: !item.isSelected } : item)),
    );
  };

  const handleSelectAll = (checked: boolean) => {
    setHistory(prevHistory => prevHistory.map(item => ({ ...item, isSelected: checked })));
  };

  const handleSave = async () => {
    const selectedItems = history.filter(item => item.isSelected);
    try {
      await client.post('/history', selectedItems);
      setHistory(prevHistory => prevHistory.map(item => ({ ...item, isSelected: false })));
      toast.success('History Saved', {
        description: <span className="font-light">Use Savvy&apos;s CLI to finish sharing your expertise</span>,
        closeButton: true,
        duration: 4000,
      });
    } catch (error: unknown) {
      const isConnectionError = isAxiosError(error) && (error.code === 'ERR_NETWORK' || error.code === 'ECONNREFUSED');
      if (isConnectionError) {
        toast.error("Can't Connect to Savvy", {
          closeButton: true,
          duration: Infinity,
          description: (
            <span className="text-pretty font-light">
              Run <span className="font-mono font-medium">savvy record history</span> in your terminal and try again.
            </span>
          ),
          action: (
            <Button
              variant="destructive"
              onClick={() => {
                navigator.clipboard.writeText('savvy record history');
                toast.dismiss();
              }}>
              <ClipboardIcon className="w-4 h-4 mr-1 inline" /> Copy Command
            </Button>
          ),
        });
      } else {
        toast.error('Error', {
          closeButton: true,
          duration: 8000,
          description:
            'An error occurred while saving your history. Please try again or contact us at support@getsavvy.so',
        });
      }
    }
  };

  return (
    <div className="flex flex-col max-h-scren">
      <div className="p-4  bg-white">
        <div className="flex items-center font-light">
          <label htmlFor="time-range" className="text-sm font-normal text-gray-700 mr-2">
            Time Range
          </label>
          <Select value={selectedHours.toString()} onValueChange={value => setSelectedHours(Number(value))}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Select time range" />
            </SelectTrigger>
            <SelectContent className="font-light">
              {TIME_RANGES.map(({ label, hours }) => (
                <SelectItem key={hours} value={hours.toString()} className="focus:text-primary">
                  Last {label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-4">
        {loading ? (
          <div className="text-gray-600">Loading history...</div>
        ) : (
          <div className="space-y-2">
            {!loading && history.length >= 2 && (
              <div className="flex items-center rounded bg-white p-3 shadow-sm hover:bg-primary/10 mb-2">
                <Checkbox
                  id="select-all"
                  checked={history.every(item => item.isSelected)}
                  onCheckedChange={handleSelectAll}
                  className="mr-2 data-[state=checked]:bg-primary/10"
                />
                <label htmlFor="select-all" className="flex-grow cursor-pointer">
                  <div className="text-sm font-light text-gray-700">Select All</div>
                </label>
              </div>
            )}
            {history.map((item, index) => (
              <div key={index} className="flex items-center rounded p-3 shadow-sm hover:bg-primary/10">
                <Checkbox
                  id={`item-${index}`}
                  checked={item.isSelected}
                  onCheckedChange={() => handleItemSelect(index)}
                  className="mr-2 data-[state=checked]:bg-primary/10"
                />
                <label htmlFor={`item-${index}`} className="flex-grow cursor-pointer">
                  <div className="text-sm font-normal text-gray-700 hover:underline-offset-2 hover:text-primary hover:underline">
                    {item.title || getHostname(item.url || '')}
                  </div>
                  <a
                    href={item.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="font-light text-xs text-gray-500 hover:underline text-ellipsis overflow-hidden">
                    {item.url ? (item.url.length > 80 ? item.url.slice(0, 80) + '...' : item.url) : ''}
                  </a>
                  <div className="text-xs font-thin text-gray-500">
                    {new Date(item.lastVisitTime!).toLocaleString()}
                  </div>
                </label>
                <a
                  href={item.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="ml-2"
                  aria-label={`Visit ${getHostname(item.url || '')}`}>
                  <Badge className="cursor-pointer font-normal bg-primary/10 text-black hover:bg-primary hover:text-white">
                    Visit <ExternalLink className="w-3 h-3 ml-1 inline" />
                  </Badge>
                </a>
              </div>
            ))}
            {history.length === 0 && (
              <div className="inline">
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger>
                      <div className="inline text-gray-600 text-sm font-medium text-pretty">
                        Savvy only shows links matching
                        <span className="font-medium underline underline-offset-2"> popular devtools.</span> Try a
                        different time range.
                      </div>
                    </TooltipTrigger>
                    <TooltipContent className="overflow-y-scroll">
                      <ScrollArea className="h-72 w-64 rounded-md">
                        <div className="p-4">
                          <ul className="list-disc ml-4 font-thin">
                            {[...allowedDomains].sort().map(domain => (
                              <li key={domain}>{domain}</li>
                            ))}
                          </ul>
                        </div>
                      </ScrollArea>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </div>
            )}
          </div>
        )}
      </div>

      {history.length > 0 && (
        <div className="p-4 bg-white sticky bottom-0">
          <Button
            onClick={handleSave}
            className="w-full bg-primary text-white"
            disabled={!history.some(item => item.isSelected)}>
            <ChevronRight className="w-4 h-4 mr-1 inline" />
            Save History
          </Button>
        </div>
      )}
      <Toaster richColors position="bottom-right" expand={true} visibleToasts={2} />
    </div>
  );
};
